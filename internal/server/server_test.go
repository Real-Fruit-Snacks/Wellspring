package server

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
	"wellspring/internal/callback"
	"wellspring/internal/payload"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return port
}

func setupTestServer(t *testing.T) (*payload.PayloadManager, *callback.Tracker, string, string) {
	t.Helper()

	manager, err := payload.NewPayloadManager()
	if err != nil {
		t.Fatal(err)
	}
	tracker := callback.NewTracker()

	httpsPort := freePort(t)
	rawPort := freePort(t)

	cfg := Config{
		ListenAddr:    "127.0.0.1:" + httpsPort,
		RawListenAddr: "127.0.0.1:" + rawPort,
		SNI:           "localhost",
	}

	srv, err := New(cfg, manager, tracker)
	if err != nil {
		t.Fatal(err)
	}

	go srv.Start()
	time.Sleep(200 * time.Millisecond)

	return manager, tracker, httpsPort, rawPort
}

func testClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func writeTestELF(t *testing.T, size int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.elf")
	elf := make([]byte, size)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[18] = 0x3E
	for i := 20; i < len(elf); i++ {
		elf[i] = byte(i % 256)
	}
	os.WriteFile(path, elf, 0755)
	return path
}

func TestPayloadDelivery(t *testing.T) {
	manager, tracker, httpsPort, _ := setupTestServer(t)

	p, err := manager.Add(writeTestELF(t, 128), "test-implant")
	if err != nil {
		t.Fatal(err)
	}

	tok, err := manager.Tokens.Generate(p.ID, time.Hour, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := testClient().Get(fmt.Sprintf("https://127.0.0.1:%s/p/%s", httpsPort, tok.Value))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) != 128 {
		t.Errorf("expected 128 bytes, got %d", len(body))
	}
	if body[0] != 0x7f || body[1] != 'E' {
		t.Error("payload data corrupted")
	}

	events := tracker.List()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if len(events) > 0 && events[0].Method != "https" {
		t.Errorf("expected https method, got %s", events[0].Method)
	}
}

func TestSingleUseToken(t *testing.T) {
	manager, _, httpsPort, _ := setupTestServer(t)

	p, _ := manager.Add(writeTestELF(t, 64), "")
	tok, _ := manager.Tokens.Generate(p.ID, time.Hour, 1, "")

	client := testClient()
	url := fmt.Sprintf("https://127.0.0.1:%s/p/%s", httpsPort, tok.Value)

	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("first request: expected 200, got %d", resp.StatusCode)
	}

	resp2, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("second request: expected 404, got %d", resp2.StatusCode)
	}
}

func TestInvalidToken(t *testing.T) {
	_, _, httpsPort, _ := setupTestServer(t)

	resp, err := testClient().Get(fmt.Sprintf("https://127.0.0.1:%s/p/nonexistent", httpsPort))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRawTLSDelivery(t *testing.T) {
	manager, tracker, _, rawPort := setupTestServer(t)

	p, err := manager.Add(writeTestELF(t, 128), "raw-test")
	if err != nil {
		t.Fatal(err)
	}

	tok, err := manager.Tokens.Generate(p.ID, time.Hour, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := tls.Dial("tcp", "127.0.0.1:"+rawPort, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send token
	fmt.Fprintf(conn, "%s\n", tok.Value)

	// Read payload
	data, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 128 {
		t.Errorf("expected 128 bytes, got %d", len(data))
	}
	if data[0] != 0x7f || data[1] != 'E' {
		t.Error("raw TLS payload data corrupted")
	}

	events := tracker.List()
	found := false
	for _, e := range events {
		if e.Method == "raw-tls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected raw-tls delivery event")
	}
}

func TestServerHeaderConsistency(t *testing.T) {
	manager, _, httpsPort, _ := setupTestServer(t)

	p, _ := manager.Add(writeTestELF(t, 64), "")
	tok, _ := manager.Tokens.Generate(p.ID, time.Hour, 0, "")

	// Payload delivery should have nginx Server header
	resp, err := testClient().Get(fmt.Sprintf("https://127.0.0.1:%s/p/%s", httpsPort, tok.Value))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.Header.Get("Server") != "nginx/1.24.0" {
		t.Errorf("payload delivery: expected nginx server header, got %q", resp.Header.Get("Server"))
	}

	// Decoy (invalid path) should also have nginx Server header
	resp2, err := testClient().Get(fmt.Sprintf("https://127.0.0.1:%s/random", httpsPort))
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()

	if resp2.Header.Get("Server") != "nginx/1.24.0" {
		t.Errorf("decoy: expected nginx server header, got %q", resp2.Header.Get("Server"))
	}
}

func TestDecoyPage(t *testing.T) {
	_, _, httpsPort, _ := setupTestServer(t)

	resp, err := testClient().Get(fmt.Sprintf("https://127.0.0.1:%s/", httpsPort))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected decoy HTML body")
	}
	if resp.Header.Get("Server") != "nginx/1.24.0" {
		t.Errorf("expected nginx server header, got %s", resp.Header.Get("Server"))
	}
}
