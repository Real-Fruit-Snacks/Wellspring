package theme

import "fmt"

// Catppuccin Mocha palette — ANSI true-color escapes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Italic  = "\033[3m"

	// Catppuccin Mocha colors
	Rosewater = "\033[38;2;245;224;220m"
	Flamingo  = "\033[38;2;242;205;205m"
	Pink      = "\033[38;2;245;194;231m"
	Mauve     = "\033[38;2;203;166;247m"
	Red       = "\033[38;2;243;139;168m"
	Maroon    = "\033[38;2;235;160;172m"
	Peach     = "\033[38;2;250;179;135m"
	Yellow    = "\033[38;2;249;226;175m"
	Green     = "\033[38;2;166;227;161m"
	Teal      = "\033[38;2;148;226;213m"
	Sky       = "\033[38;2;137;220;235m"
	Sapphire  = "\033[38;2;116;199;236m"
	Blue      = "\033[38;2;137;180;250m"
	Lavender  = "\033[38;2;180;190;254m"

	Text     = "\033[38;2;205;214;244m"
	Subtext1 = "\033[38;2;186;194;222m"
	Subtext0 = "\033[38;2;166;173;200m"
	Overlay2 = "\033[38;2;147;153;178m"
	Overlay1 = "\033[38;2;127;132;156m"
	Overlay0 = "\033[38;2;108;112;134m"
	Surface2 = "\033[38;2;88;91;112m"
	Surface1 = "\033[38;2;69;71;90m"
	Surface0 = "\033[38;2;49;50;68m"
	Base     = "\033[38;2;30;30;46m"
	Mantle   = "\033[38;2;24;24;37m"
	Crust    = "\033[38;2;17;17;27m"
)

// Shorthand log helpers
func Info(format string, a ...any) string {
	return fmt.Sprintf("%s[*]%s %s", Blue, Reset, fmt.Sprintf(format, a...))
}

func Success(format string, a ...any) string {
	return fmt.Sprintf("%s[+]%s %s", Green, Reset, fmt.Sprintf(format, a...))
}

func Warn(format string, a ...any) string {
	return fmt.Sprintf("%s[!]%s %s", Yellow, Reset, fmt.Sprintf(format, a...))
}

func Error(format string, a ...any) string {
	return fmt.Sprintf("%s[-]%s %s", Red, Reset, fmt.Sprintf(format, a...))
}
