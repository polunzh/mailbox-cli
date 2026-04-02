package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// CallbackServer listens on a random local port for OAuth callback redirects.
type CallbackServer struct {
	port     int
	codeCh   chan string
	listener net.Listener
}

// NewCallbackServer starts a local HTTP server on a random port.
func NewCallbackServer() (*CallbackServer, error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("oauth callback: listen: %w", err)
	}
	srv := &CallbackServer{
		port:     ln.Addr().(*net.TCPAddr).Port,
		codeCh:   make(chan string, 1),
		listener: ln,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		_, _ = fmt.Fprintln(w, "Authentication complete. You may close this tab.")
		srv.codeCh <- code
	})

	go http.Serve(ln, mux) //nolint:errcheck
	return srv, nil
}

// Port returns the port the server is listening on.
func (s *CallbackServer) Port() int {
	return s.port
}

// WaitForCode blocks until the callback delivers a code or the timeout elapses.
func (s *CallbackServer) WaitForCode(timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	defer func() {
		_ = s.listener.Close()
	}()

	select {
	case code := <-s.codeCh:
		return code, nil
	case <-ctx.Done():
		return "", fmt.Errorf("oauth callback: timed out waiting for authorization code")
	}
}
