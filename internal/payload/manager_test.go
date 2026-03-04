package payload

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "implant.elf")
	elf := make([]byte, 256)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[18] = 0x3E
	// Fill with recognizable pattern
	for i := 20; i < len(elf); i++ {
		elf[i] = byte(i)
	}
	os.WriteFile(path, elf, 0755)
	return path
}

func TestManagerAddAndGet(t *testing.T) {
	pm, err := NewPayloadManager()
	if err != nil {
		t.Fatal(err)
	}

	path := writeTestBinary(t)
	p, err := pm.Add(path, "")
	if err != nil {
		t.Fatal(err)
	}

	if p.ID != "p1" {
		t.Errorf("expected p1, got %s", p.ID)
	}
	if !p.Encrypted {
		t.Error("payload should be encrypted")
	}
	if p.Arch.OS != "linux" || p.Arch.Arch != "amd64" {
		t.Errorf("expected linux/amd64, got %s/%s", p.Arch.OS, p.Arch.Arch)
	}

	got, ok := pm.Get(p.ID)
	if !ok {
		t.Fatal("payload not found")
	}
	if got.Name != "implant.elf" {
		t.Errorf("expected implant.elf, got %s", got.Name)
	}
}

func TestManagerAddWithName(t *testing.T) {
	pm, _ := NewPayloadManager()
	path := writeTestBinary(t)
	p, _ := pm.Add(path, "custom-name")
	if p.Name != "custom-name" {
		t.Errorf("expected custom-name, got %s", p.Name)
	}
}

func TestManagerGetDecrypted(t *testing.T) {
	pm, _ := NewPayloadManager()
	path := writeTestBinary(t)
	p, _ := pm.Add(path, "")

	data, err := pm.GetDecrypted(p.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 256 {
		t.Errorf("expected 256 bytes, got %d", len(data))
	}

	// Verify ELF header survived round-trip
	if data[0] != 0x7f || data[1] != 'E' || data[2] != 'L' || data[3] != 'F' {
		t.Error("ELF header corrupted after decrypt")
	}
}

func TestManagerRemove(t *testing.T) {
	pm, _ := NewPayloadManager()
	path := writeTestBinary(t)
	p, _ := pm.Add(path, "")

	if !pm.Remove(p.ID) {
		t.Error("remove should return true")
	}
	if pm.Remove(p.ID) {
		t.Error("second remove should return false")
	}
	if _, ok := pm.Get(p.ID); ok {
		t.Error("payload should be gone after remove")
	}
}

func TestManagerList(t *testing.T) {
	pm, _ := NewPayloadManager()
	path := writeTestBinary(t)
	pm.Add(path, "one")
	pm.Add(path, "two")

	list := pm.List()
	if len(list) != 2 {
		t.Errorf("expected 2 payloads, got %d", len(list))
	}
}

func TestManagerZeroAll(t *testing.T) {
	pm, _ := NewPayloadManager()
	path := writeTestBinary(t)
	pm.Add(path, "")

	pm.ZeroAll()

	list := pm.List()
	if len(list) != 0 {
		t.Errorf("expected 0 payloads after ZeroAll, got %d", len(list))
	}
}
