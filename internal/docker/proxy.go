package docker

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type Proxy struct {
	allowlist []string
	denylist  []string
	server    *http.Server
}

func domainMatch(pattern, host string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:]
		if host == suffix[1:] {
			return true
		}
		if strings.HasSuffix(host, suffix) {
			return true
		}
		return false
	}
	return pattern == host
}

func (p *Proxy) isAllowed(host string) bool {
	if len(p.allowlist) > 0 {
		for _, pattern := range p.allowlist {
			if domainMatch(pattern, host) {
				return true
			}
		}
		return false
	}
	for _, pattern := range p.denylist {
		if domainMatch(pattern, host) {
			return false
		}
	}
	return true
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.handleConnect(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Hostname()
	if !p.isAllowed(host) {
		slog.Info("request denied by network policy", "host", host)
		http.Error(w, fmt.Sprintf("Access denied by network policy: %s", host), http.StatusForbidden)
		return
	}
	slog.Debug("request allowed", "host", host)

	transport := &http.Transport{}
	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	if !p.isAllowed(host) {
		slog.Info("CONNECT denied by network policy", "host", host)
		http.Error(w, fmt.Sprintf("Access denied by network policy: %s", host), http.StatusForbidden)
		return
	}
	slog.Debug("CONNECT allowed", "host", host)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	proxyConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer proxyConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(proxyConn, clientConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, proxyConn)
		done <- struct{}{}
	}()
	<-done
}

func RunProxy(allowlist, denylist []string, bindAddr string) (*Proxy, int, error) {
	p := &Proxy{
		allowlist: allowlist,
		denylist:  denylist,
	}

	listener, err := net.Listen("tcp", bindAddr+":0")
	if err != nil {
		return nil, 0, fmt.Errorf("starting proxy listener on %s: %w", bindAddr, err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	p.server = &http.Server{
		Handler: p,
	}

	go p.server.Serve(listener)

	slog.Info("proxy started", "addr", fmt.Sprintf("%s:%d", bindAddr, port))

	return p, port, nil
}

func (p *Proxy) Close() {
	if p.server != nil {
		p.server.Close()
	}
	slog.Info("proxy stopped")
}
