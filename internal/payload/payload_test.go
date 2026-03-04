package payload

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectArch_ELF_AMD64(t *testing.T) {
	// Minimal ELF header for x86-64
	elf := make([]byte, 64)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[18] = 0x3E // EM_X86_64 (little-endian)
	elf[19] = 0x00

	info := DetectArch(elf)
	if info.OS != "linux" {
		t.Errorf("expected linux, got %s", info.OS)
	}
	if info.Arch != "amd64" {
		t.Errorf("expected amd64, got %s", info.Arch)
	}
}

func TestDetectArch_ELF_ARM64(t *testing.T) {
	elf := make([]byte, 64)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[18] = 0xB7 // EM_AARCH64
	elf[19] = 0x00

	info := DetectArch(elf)
	if info.OS != "linux" || info.Arch != "arm64" {
		t.Errorf("expected linux/arm64, got %s/%s", info.OS, info.Arch)
	}
}

func TestDetectArch_PE_AMD64(t *testing.T) {
	pe := make([]byte, 0x80)
	pe[0], pe[1] = 'M', 'Z'
	// PE offset at 0x3C pointing to 0x40
	pe[0x3C] = 0x40
	// PE signature
	pe[0x40], pe[0x41] = 'P', 'E'
	pe[0x44] = 0x64 // IMAGE_FILE_MACHINE_AMD64
	pe[0x45] = 0x86

	info := DetectArch(pe)
	if info.OS != "windows" {
		t.Errorf("expected windows, got %s", info.OS)
	}
	if info.Arch != "amd64" {
		t.Errorf("expected amd64, got %s", info.Arch)
	}
}

func TestDetectArch_Unknown(t *testing.T) {
	info := DetectArch([]byte("not a binary"))
	if info.OS != "unknown" || info.Arch != "unknown" {
		t.Errorf("expected unknown/unknown, got %s/%s", info.OS, info.Arch)
	}
}

func TestDetectArch_TooShort(t *testing.T) {
	info := DetectArch([]byte{0x7f})
	if info.OS != "unknown" {
		t.Errorf("expected unknown for short input, got %s", info.OS)
	}
}

func TestDetectArch_PE_MaliciousOffset(t *testing.T) {
	pe := make([]byte, 0x80)
	pe[0], pe[1] = 'M', 'Z'
	// Set PE offset to maximum uint32 value — should not panic
	pe[0x3C] = 0xFF
	pe[0x3D] = 0xFF
	pe[0x3E] = 0xFF
	pe[0x3F] = 0xFF

	info := DetectArch(pe)
	if info.OS != "windows" {
		t.Errorf("expected windows, got %s", info.OS)
	}
	// Arch should be unknown since offset is out of bounds
	if info.Arch != "unknown" {
		t.Errorf("expected unknown arch for OOB offset, got %s", info.Arch)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("test payload data that should survive round-trip encryption")
	ct, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	if string(ct) == string(plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	dec, err := Decrypt(ct, key)
	if err != nil {
		t.Fatal(err)
	}

	if string(dec) != string(plaintext) {
		t.Errorf("decrypted data doesn't match: got %q", dec)
	}
}

func TestEncryptDecrypt_WrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	ct, _ := Encrypt([]byte("secret"), key1)
	_, err := Decrypt(ct, key2)
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")

	// Write a fake ELF
	elf := make([]byte, 64)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[18] = 0x3E
	os.WriteFile(path, elf, 0755)

	p, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if p.Name != "test.bin" {
		t.Errorf("expected name test.bin, got %s", p.Name)
	}
	if p.Arch.OS != "linux" || p.Arch.Arch != "amd64" {
		t.Errorf("expected linux/amd64, got %s/%s", p.Arch.OS, p.Arch.Arch)
	}
	if p.Size != 64 {
		t.Errorf("expected size 64, got %d", p.Size)
	}
}
