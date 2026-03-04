package loader

import "fmt"

type StagerOpts struct {
	Host       string
	Port       string
	Token      string
	PayloadID  string
	TargetOS   string
	TargetArch string
	TLS        bool
}

type Loader interface {
	Name() string
	Description() string
	Generate(opts StagerOpts) (string, error)
	Supports(os, arch string) bool
	RequiresTool() string
}

// ValidateOpts checks that host, port, and token contain only shell-safe characters.
func ValidateOpts(opts StagerOpts) error {
	if err := validateShellSafe("host", opts.Host, true); err != nil {
		return err
	}
	if err := validateShellSafe("port", opts.Port, false); err != nil {
		return err
	}
	if err := validateShellSafe("token", opts.Token, false); err != nil {
		return err
	}
	return nil
}

func validateShellSafe(field, s string, allowColon bool) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_' {
			continue
		}
		if c == ':' && allowColon {
			continue
		}
		return fmt.Errorf("unsafe character %q in %s", c, field)
	}
	return nil
}
