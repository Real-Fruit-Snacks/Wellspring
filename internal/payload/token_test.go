package payload

import (
	"testing"
	"time"
)

func TestTokenGenerate(t *testing.T) {
	store := NewTokenStore()
	tok, err := store.Generate("p1", time.Hour, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tok.Value) != 32 { // 16 bytes hex-encoded
		t.Errorf("expected 32-char token, got %d", len(tok.Value))
	}
	if tok.PayloadID != "p1" {
		t.Errorf("expected payload p1, got %s", tok.PayloadID)
	}
}

func TestTokenValidate(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Hour, 0, "")

	got, ok := store.Validate(tok.Value, "1.2.3.4:1234")
	if !ok {
		t.Fatal("expected valid token")
	}
	if got.PayloadID != "p1" {
		t.Errorf("expected p1, got %s", got.PayloadID)
	}
}

func TestTokenSingleUse(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Hour, 1, "")

	_, ok := store.Validate(tok.Value, "1.2.3.4:1234")
	if !ok {
		t.Fatal("first use should succeed")
	}

	_, ok = store.Validate(tok.Value, "1.2.3.4:1234")
	if ok {
		t.Error("second use of single-use token should fail")
	}
}

func TestTokenExpired(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Millisecond, 0, "")

	time.Sleep(5 * time.Millisecond)

	_, ok := store.Validate(tok.Value, "1.2.3.4:1234")
	if ok {
		t.Error("expired token should be invalid")
	}
}

func TestTokenSourceLock_IP(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Hour, 0, "10.0.0.5")

	_, ok := store.Validate(tok.Value, "10.0.0.5:4444")
	if !ok {
		t.Error("matching IP should be allowed")
	}

	tok2, _ := store.Generate("p1", time.Hour, 0, "10.0.0.5")
	_, ok = store.Validate(tok2.Value, "10.0.0.6:4444")
	if ok {
		t.Error("non-matching IP should be rejected")
	}
}

func TestTokenSourceLock_CIDR(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Hour, 0, "10.0.0.0/24")

	_, ok := store.Validate(tok.Value, "10.0.0.99:4444")
	if !ok {
		t.Error("IP within CIDR should be allowed")
	}

	tok2, _ := store.Generate("p1", time.Hour, 0, "10.0.0.0/24")
	_, ok = store.Validate(tok2.Value, "10.0.1.1:4444")
	if ok {
		t.Error("IP outside CIDR should be rejected")
	}
}

func TestTokenInvalid(t *testing.T) {
	store := NewTokenStore()
	_, ok := store.Validate("nonexistent", "1.2.3.4:1234")
	if ok {
		t.Error("nonexistent token should be invalid")
	}
}

func TestTokenRevoke(t *testing.T) {
	store := NewTokenStore()
	tok, _ := store.Generate("p1", time.Hour, 0, "")

	if !store.Revoke(tok.Value) {
		t.Error("revoke should return true for existing token")
	}

	_, ok := store.Validate(tok.Value, "1.2.3.4:1234")
	if ok {
		t.Error("revoked token should be invalid")
	}

	if store.Revoke("nonexistent") {
		t.Error("revoke should return false for nonexistent token")
	}
}

func TestTokenPurgeExpired(t *testing.T) {
	store := NewTokenStore()
	store.Generate("p1", time.Millisecond, 0, "")
	store.Generate("p2", time.Hour, 0, "")

	time.Sleep(5 * time.Millisecond)
	count := store.PurgeExpired()
	if count != 1 {
		t.Errorf("expected 1 purged, got %d", count)
	}

	remaining := store.List()
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(remaining))
	}
}
