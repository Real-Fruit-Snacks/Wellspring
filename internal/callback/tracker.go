package callback

import (
	"net"
	"sync"
	"time"
)

type DeliveryEvent struct {
	Timestamp   time.Time
	PayloadID   string
	PayloadName string
	Token       string
	RemoteAddr  string
	Method      string // "https" or "raw-tls"
	Size        int64
	Success     bool
}

type Tracker struct {
	mu     sync.RWMutex
	events []DeliveryEvent
}

func NewTracker() *Tracker {
	return &Tracker{}
}

const maxEvents = 10000

func (t *Tracker) Record(e DeliveryEvent) {
	e.Timestamp = time.Now()
	if host, _, err := net.SplitHostPort(e.RemoteAddr); err == nil {
		e.RemoteAddr = host
	}
	t.mu.Lock()
	t.events = append(t.events, e)
	if len(t.events) > maxEvents {
		compacted := make([]DeliveryEvent, maxEvents)
		copy(compacted, t.events[len(t.events)-maxEvents:])
		t.events = compacted
	}
	t.mu.Unlock()
}

func (t *Tracker) List() []DeliveryEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]DeliveryEvent, len(t.events))
	copy(result, t.events)
	return result
}
