package payload

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type PayloadManager struct {
	mu       sync.RWMutex
	payloads map[string]*Payload
	key      []byte
	nextID   atomic.Int64
	Tokens   *TokenStore
}

func NewPayloadManager() (*PayloadManager, error) {
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	return &PayloadManager{
		payloads: make(map[string]*Payload),
		key:      key,
		Tokens:   NewTokenStore(),
	}, nil
}

func (pm *PayloadManager) Add(path string, name string) (*Payload, error) {
	p, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	if name != "" {
		p.Name = name
	}

	// Encrypt the payload data at rest
	encrypted, err := Encrypt(p.Data, pm.key)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}

	// Zero the plaintext
	for i := range p.Data {
		p.Data[i] = 0
	}
	runtime.KeepAlive(p.Data)

	p.Data = encrypted
	p.Encrypted = true
	p.ID = fmt.Sprintf("p%d", pm.nextID.Add(1))

	pm.mu.Lock()
	pm.payloads[p.ID] = p
	pm.mu.Unlock()

	return p, nil
}

func (pm *PayloadManager) Get(id string) (*Payload, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.payloads[id]
	return p, ok
}

func (pm *PayloadManager) GetDecrypted(id string) ([]byte, error) {
	pm.mu.RLock()
	p, ok := pm.payloads[id]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("payload %s not found", id)
	}

	return Decrypt(p.Data, pm.key)
}

func (pm *PayloadManager) Remove(id string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, ok := pm.payloads[id]
	if !ok {
		return false
	}

	// Zero memory
	for i := range p.Data {
		p.Data[i] = 0
	}
	runtime.KeepAlive(p.Data)

	delete(pm.payloads, id)
	return true
}

func (pm *PayloadManager) List() []*Payload {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*Payload, 0, len(pm.payloads))
	for _, p := range pm.payloads {
		result = append(result, p)
	}
	return result
}

func (pm *PayloadManager) ZeroAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range pm.payloads {
		for i := range p.Data {
			p.Data[i] = 0
		}
		runtime.KeepAlive(p.Data)
	}
	for i := range pm.key {
		pm.key[i] = 0
	}
	runtime.KeepAlive(pm.key)
	pm.Tokens.ZeroAll()
	pm.payloads = make(map[string]*Payload)
}
