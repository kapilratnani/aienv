package docker

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kapilratnani/aienv/internal/audit"
)

type Proxy struct {
	allowlist   []string
	denylist    []string
	server      *http.Server
	learn       bool
	auditWriter *audit.Writer
	learnHosts  map[string]bool
	mu          sync.Mutex
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

func (p *Proxy) record(host, method string, allowed bool) {
	if p.learn {
		p.mu.Lock()
		p.learnHosts[host] = true
		p.mu.Unlock()
	}
	if p.auditWriter != nil {
		p.auditWriter.AppendNetworkEntry(audit.NetworkEntry{
			Timestamp: time.Now().UTC(),
			Host:      host,
			Method:    method,
			Allowed:   allowed,
		})
	}
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
	allowed := p.isAllowed(host)
	p.record(host, r.Method, allowed)

	if !allowed {
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
	allowed := p.isAllowed(host)
	p.record(host, "CONNECT", allowed)

	if !allowed {
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

func RunProxy(allowlist, denylist []string, bindAddr string, auditWriter *audit.Writer) (*Proxy, int, error) {
	learn := len(allowlist) == 0 && len(denylist) == 0

	p := &Proxy{
		allowlist:   allowlist,
		denylist:    denylist,
		learn:       learn,
		auditWriter: auditWriter,
		learnHosts:  make(map[string]bool),
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

	slog.Info("proxy started", "addr", fmt.Sprintf("%s:%d", bindAddr, port), "learn", learn)

	return p, port, nil
}

func (p *Proxy) Close() {
	if p.server != nil {
		p.server.Close()
	}
	slog.Info("proxy stopped")

	// Print learn mode summary to stderr
	if p.learn && len(p.learnHosts) > 0 {
		hosts := make([]string, 0, len(p.learnHosts))
		for h := range p.learnHosts {
			hosts = append(hosts, h)
		}
		sort.Strings(hosts)
		fmt.Fprintf(os.Stderr, "\n--- Learn Mode: Hosts Accessed ---\n")
		for _, h := range hosts {
			fmt.Fprintf(os.Stderr, "  %s\n", h)
		}
		fmt.Fprintf(os.Stderr, "---\n")
		fmt.Fprintf(os.Stderr, "Add these to permissions.network.allow in your env config.\n\n")
	}
}
