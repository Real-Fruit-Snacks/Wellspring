package payload

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type Token struct {
	Value      string
	PayloadID  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	MaxUses    int
	UseCount   int
	SourceLock string // CIDR or IP
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *Token) IsExhausted() bool {
	return t.MaxUses > 0 && t.UseCount >= t.MaxUses
}

func (t *Token) IsSourceAllowed(remoteAddr string) bool {
	if t.SourceLock == "" {
		return true
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	// Try CIDR match
	_, cidr, err := net.ParseCIDR(t.SourceLock)
	if err == nil {
		ip := net.ParseIP(host)
		return ip != nil && cidr.Contains(ip)
	}

	// Direct IP match
	return host == t.SourceLock
}

type TokenStore struct {
	mu      sync.RWMutex
	tokens  map[string]*Token // keyed by HMAC hash, not raw value
	hmacKey []byte
}

func NewTokenStore() *TokenStore {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic("wellspring: failed to initialize HMAC key: " + err.Error())
	}
	return &TokenStore{
		tokens:  make(map[string]*Token),
		hmacKey: key,
	}
}

// hashToken produces an HMAC-SHA256 digest of the token value.
// Using this as the map key prevents timing attacks on map lookup.
func (s *TokenStore) hashToken(value string) string {
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *TokenStore) Generate(payloadID string, ttl time.Duration, maxUses int, sourceLock string) (*Token, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	token := &Token{
		Value:      hex.EncodeToString(b),
		PayloadID:  payloadID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		MaxUses:    maxUses,
		SourceLock: sourceLock,
	}

	key := s.hashToken(token.Value)

	s.mu.Lock()
	s.tokens[key] = token
	s.mu.Unlock()

	return token, nil
}

func (s *TokenStore) Validate(value, remoteAddr string) (*Token, bool) {
	key := s.hashToken(value)

	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokens[key]
	if !ok {
		return nil, false
	}

	if token.IsExpired() || token.IsExhausted() {
		delete(s.tokens, key)
		return nil, false
	}

	if !token.IsSourceAllowed(remoteAddr) {
		return nil, false
	}

	token.UseCount++
	return token, true
}

func (s *TokenStore) Revoke(value string) bool {
	key := s.hashToken(value)

	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tokens[key]
	if ok {
		zeroString(&t.Value)
		zeroString(&t.SourceLock)
		delete(s.tokens, key)
	}
	return ok
}

func (s *TokenStore) List() []*Token {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Token, 0, len(s.tokens))
	for _, t := range s.tokens {
		result = append(result, t)
	}
	return result
}

func (s *TokenStore) PurgeExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for k, t := range s.tokens {
		if t.IsExpired() || t.IsExhausted() {
			zeroString(&t.Value)
			zeroString(&t.SourceLock)
			delete(s.tokens, k)
			count++
		}
	}
	return count
}

func (s *TokenStore) ZeroAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, t := range s.tokens {
		zeroString(&t.Value)
		zeroString(&t.SourceLock)
		delete(s.tokens, k)
	}
	for i := range s.hmacKey {
		s.hmacKey[i] = 0
	}
	runtime.KeepAlive(s.hmacKey)
}

// zeroString overwrites the backing array of a string with zeros in-place.
// Uses unsafe.StringData (Go 1.20+) to access the actual backing memory
// rather than a copy. Safe for heap-allocated strings (e.g. hex-encoded tokens).
func zeroString(s *string) {
	if len(*s) == 0 {
		return
	}
	p := unsafe.StringData(*s)
	b := unsafe.Slice(p, len(*s))
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
	*s = ""
}
