package payload

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type ArchInfo struct {
	OS   string // "linux", "windows", "unknown"
	Arch string // "amd64", "386", "arm64", "unknown"
}

type Payload struct {
	ID        string
	Name      string
	Data      []byte // encrypted at rest
	Size      int64
	Arch      ArchInfo
	Encrypted bool
}

// DetectArch identifies OS and architecture from ELF/PE headers.
func DetectArch(data []byte) ArchInfo {
	if len(data) < 64 {
		return ArchInfo{OS: "unknown", Arch: "unknown"}
	}

	// ELF: 0x7f 'E' 'L' 'F'
	if data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		info := ArchInfo{OS: "linux", Arch: "unknown"}
		if len(data) > 19 {
			machine := binary.LittleEndian.Uint16(data[18:20])
			switch machine {
			case 0x3E:
				info.Arch = "amd64"
			case 0x03:
				info.Arch = "386"
			case 0xB7:
				info.Arch = "arm64"
			case 0x28:
				info.Arch = "arm"
			default:
				info.Arch = "unknown"
			}
		}
		return info
	}

	// PE: 'M' 'Z'
	if data[0] == 'M' && data[1] == 'Z' {
		info := ArchInfo{OS: "windows", Arch: "unknown"}
		if len(data) > 0x3F {
			peOffset := binary.LittleEndian.Uint32(data[0x3C:0x40])
			if peOffset >= 4 && int64(peOffset)+6 <= int64(len(data)) && data[peOffset] == 'P' && data[peOffset+1] == 'E' {
				machine := binary.LittleEndian.Uint16(data[peOffset+4 : peOffset+6])
				switch machine {
				case 0x8664:
					info.Arch = "amd64"
				case 0x014c:
					info.Arch = "386"
				case 0xAA64:
					info.Arch = "arm64"
				default:
					info.Arch = "unknown"
				}
			}
		}
		return info
	}

	return ArchInfo{OS: "unknown", Arch: "unknown"}
}

const MaxPayloadSize = 100 << 20 // 100MB

func LoadFromFile(path string) (*Payload, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat payload: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file")
	}
	if info.Size() > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large (%d bytes, max %d)", info.Size(), MaxPayloadSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	arch := DetectArch(data)

	return &Payload{
		Name: filepath.Base(path),
		Data: data,
		Size: int64(len(data)),
		Arch: arch,
	}, nil
}
