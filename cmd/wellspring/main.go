package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"wellspring/internal/antiforensics"
	"wellspring/internal/callback"
	"wellspring/internal/cli"
	"wellspring/internal/payload"
	"wellspring/internal/server"
	"wellspring/internal/theme"
)

func main() {
	listen := flag.String("listen", ":443", "HTTPS listener address")
	rawListen := flag.String("raw-listen", ":4443", "Raw TLS listener address")
	certPath := flag.String("cert", "", "TLS certificate path")
	keyPath := flag.String("key", "", "TLS private key path")
	sni := flag.String("sni", "cloudflare-dns.com", "SNI/CN for generated cert")
	decoy := flag.String("decoy", "", "Path to HTML decoy page")
	genCert := flag.Bool("gencert", false, "Generate TLS cert and exit")
	flag.Parse()

	if *genCert {
		certFile := "server.crt"
		keyFile := "server.key"
		if *certPath != "" {
			certFile = *certPath
		}
		if *keyPath != "" {
			keyFile = *keyPath
		}
		if err := server.GenerateCertToFile(certFile, keyFile, *sni); err != nil {
			fmt.Fprintln(os.Stderr, theme.Error("%v", err))
			os.Exit(1)
		}
		fmt.Println(theme.Success("Certificate: %s%s%s", theme.Teal, certFile, theme.Reset))
		fmt.Println(theme.Success("Key:         %s%s%s", theme.Teal, keyFile, theme.Reset))
		fmt.Println(theme.Success("CN/SAN:      %s%s%s", theme.Teal, *sni, theme.Reset))
		return
	}

	manager, err := payload.NewPayloadManager()
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error("Init failed: %v", err))
		os.Exit(1)
	}

	tracker := callback.NewTracker()

	enforcer := antiforensics.NewExpiryEnforcer(manager, 30*time.Second)
	enforcer.Start()
	defer enforcer.Stop()

	// Signal handler: zero all secrets on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Printf("\n%s[*] Signal received — zeroing all payloads and keys...%s\n", theme.Yellow, theme.Reset)
		enforcer.Stop()
		manager.ZeroAll()
		os.Exit(0)
	}()

	cfg := server.Config{
		ListenAddr:    *listen,
		RawListenAddr: *rawListen,
		CertPath:      *certPath,
		KeyPath:       *keyPath,
		SNI:           *sni,
		DecoyPath:     *decoy,
	}

	srv, err := server.New(cfg, manager, tracker)
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error("Server init failed: %v", err))
		os.Exit(1)
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Start()
	}()

	// Brief pause to catch immediate bind errors
	time.Sleep(100 * time.Millisecond)
	select {
	case err := <-srvErr:
		fmt.Fprintln(os.Stderr, theme.Error("Server error: %v", err))
		os.Exit(1)
	default:
	}

	// Extract host/port for CLI display
	host := "0.0.0.0"
	port := extractPort(*listen)
	rawPort := extractPort(*rawListen)

	console := cli.NewConsole(manager, tracker, host, port, rawPort)
	console.Run()
}

func extractPort(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		return addr[idx+1:]
	}
	return addr
}
