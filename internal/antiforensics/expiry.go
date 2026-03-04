package antiforensics

import (
	"runtime"
	"sync"
	"time"
	"wellspring/internal/payload"
)

type ExpiryEnforcer struct {
	manager  *payload.PayloadManager
	interval time.Duration
	stop     chan struct{}
	stopOnce sync.Once
}

func NewExpiryEnforcer(manager *payload.PayloadManager, interval time.Duration) *ExpiryEnforcer {
	return &ExpiryEnforcer{
		manager:  manager,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (e *ExpiryEnforcer) Start() {
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.manager.Tokens.PurgeExpired()
			case <-e.stop:
				return
			}
		}
	}()
}

func (e *ExpiryEnforcer) Stop() {
	e.stopOnce.Do(func() { close(e.stop) })
}

func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
}
