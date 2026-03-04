package antiforensics

import (
	"testing"
	"time"
	"wellspring/internal/payload"
)

func TestZeroBytes(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03}
	ZeroBytes(data)
	for i, b := range data {
		if b != 0 {
			t.Errorf("byte %d not zeroed: got 0x%02x", i, b)
		}
	}
}

func TestZeroBytesEmpty(t *testing.T) {
	// Should not panic on empty/nil slices
	ZeroBytes(nil)
	ZeroBytes([]byte{})
}

func TestExpiryEnforcer(t *testing.T) {
	manager, err := payload.NewPayloadManager()
	if err != nil {
		t.Fatal(err)
	}

	// Generate a token with very short TTL
	_, err = manager.Tokens.Generate("p1", time.Millisecond, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	// Generate a long-lived token
	_, err = manager.Tokens.Generate("p2", time.Hour, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	enforcer := NewExpiryEnforcer(manager, 10*time.Millisecond)
	enforcer.Start()
	defer enforcer.Stop()

	// Wait for at least one purge cycle
	time.Sleep(50 * time.Millisecond)

	tokens := manager.Tokens.List()
	if len(tokens) != 1 {
		t.Errorf("expected 1 remaining token after purge, got %d", len(tokens))
	}
}

func TestExpiryEnforcerDoubleStop(t *testing.T) {
	manager, _ := payload.NewPayloadManager()
	enforcer := NewExpiryEnforcer(manager, time.Second)
	enforcer.Start()
	// Double stop should not panic (sync.Once)
	enforcer.Stop()
	enforcer.Stop()
}
