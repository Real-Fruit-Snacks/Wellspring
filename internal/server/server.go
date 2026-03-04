package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
	"wellspring/internal/callback"
	"wellspring/internal/payload"
	"wellspring/internal/theme"
)

const maxRawTLSConcurrent = 128

type Config struct {
	ListenAddr    string
	RawListenAddr string
	CertPath      string
	KeyPath       string
	SNI           string
	DecoyPath     string
}

type WellspringServer struct {
	config    Config
	tlsConfig *tls.Config
	manager   *payload.PayloadManager
	tracker   *callback.Tracker
	decoyHTML []byte
}

func New(cfg Config, manager *payload.PayloadManager, tracker *callback.Tracker) (*WellspringServer, error) {
	tlsCfg, err := NewTLSConfig(cfg.CertPath, cfg.KeyPath, cfg.SNI)
	if err != nil {
		return nil, fmt.Errorf("TLS setup: %w", err)
	}

	var decoy []byte
	if cfg.DecoyPath != "" {
		var readErr error
		decoy, readErr = os.ReadFile(cfg.DecoyPath)
		if readErr != nil {
			fmt.Fprintln(os.Stderr, theme.Warn("Failed to load decoy page %s: %v", cfg.DecoyPath, readErr))
		}
	}
	if decoy == nil {
		decoy = defaultDecoyHTML()
	}

	return &WellspringServer{
		config:    cfg,
		tlsConfig: tlsCfg,
		manager:   manager,
		tracker:   tracker,
		decoyHTML: decoy,
	}, nil
}

func (s *WellspringServer) Start() error {
	errCh := make(chan error, 2)

	// HTTPS listener
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/p/", s.handlePayloadDelivery)
		mux.HandleFunc("/", s.handleDecoy)

		ln, err := tls.Listen("tcp", s.config.ListenAddr, s.tlsConfig)
		if err != nil {
			errCh <- fmt.Errorf("HTTPS listen %s: %w", s.config.ListenAddr, err)
			return
		}
		fmt.Println(theme.Info("HTTPS listener on %s%s%s", theme.Teal, s.config.ListenAddr, theme.Reset))
		srv := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    64 << 10, // 64KB
		}
		errCh <- srv.Serve(ln)
	}()

	// Raw TLS listener
	go func() {
		rawTLSCfg := s.tlsConfig.Clone()
		rawTLSCfg.NextProtos = nil // raw listener is not HTTP — don't advertise ALPN
		ln, err := tls.Listen("tcp", s.config.RawListenAddr, rawTLSCfg)
		if err != nil {
			errCh <- fmt.Errorf("raw TLS listen %s: %w", s.config.RawListenAddr, err)
			return
		}
		fmt.Println(theme.Info("Raw TLS listener on %s%s%s", theme.Teal, s.config.RawListenAddr, theme.Reset))
		sem := make(chan struct{}, maxRawTLSConcurrent)
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			sem <- struct{}{}
			go func() {
				defer func() { <-sem }()
				s.handleRawTLS(conn)
			}()
		}
	}()

	return <-errCh
}
