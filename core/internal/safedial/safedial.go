// Package safedial provides an HTTP transport for fetching untrusted URLs
// that refuses loopback, private, and otherwise non-public network addresses,
// including across redirects and DNS rebinding.
package safedial

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

// AllowPrivate disables the public-address checks so tests can reach local
// httptest servers. Production code must leave it false.
var AllowPrivate atomic.Bool

// maxRedirects bounds the redirect chain of one request.
const maxRedirects = 4

// Transport returns an HTTP transport that dials only public addresses.
func Transport() *http.Transport {
	return &http.Transport{
		DialContext:           DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       time.Minute,
		MaxIdleConns:          16,
		MaxIdleConnsPerHost:   2,
	}
}

// CheckRedirect limits redirects and re-checks each target address, so a
// redirect cannot escape to a private network or downgrade from https.
func CheckRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return errors.New("too many redirects")
	}
	if err := Validate(request.URL.String()); err != nil {
		return err
	}
	if len(via) > 0 && via[0].URL.Scheme == "https" && request.URL.Scheme != "https" {
		return errors.New("redirects may not downgrade to http")
	}
	return nil
}

// DialContext resolves the request host once, rejects unsafe addresses, and
// dials the validated IP so DNS cannot change between checking and connecting.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, errors.New("only TCP requests are allowed")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || port == "" {
		return nil, errors.New("invalid request host")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, errors.New("invalid request port")
	}

	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	selected := make([]net.IP, 0, len(addresses))
	for _, candidate := range addresses {
		if !AllowPrivate.Load() && !PublicAddress(candidate.IP) {
			return nil, errors.New("the host resolves to a forbidden address")
		}
		if network == "tcp" ||
			network == "tcp4" && candidate.IP.To4() != nil ||
			network == "tcp6" && candidate.IP.To4() == nil {
			selected = append(selected, candidate.IP)
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("the host has no usable address")
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return dialer.DialContext(ctx, network, net.JoinHostPort(selected[0].String(), port))
}

// PublicAddress rejects loopback, private, shared, link-local, multicast, and
// unspecified addresses.
func PublicAddress(address net.IP) bool {
	if address.IsLoopback() || address.IsPrivate() || address.IsLinkLocalUnicast() ||
		address.IsLinkLocalMulticast() || address.IsMulticast() || address.IsUnspecified() {
		return false
	}
	if address.To4() == nil {
		return true
	}
	for _, blocked := range []net.IPNet{
		{IP: net.IPv4(100, 64, 0, 0).To4(), Mask: net.CIDRMask(10, 32)},
		{IP: net.IPv4(198, 18, 0, 0).To4(), Mask: net.CIDRMask(15, 32)},
	} {
		if blocked.Contains(address) {
			return false
		}
	}
	return true
}

// Validate applies the protocol-level check to one request address, including
// literal IPs, and is re-applied after redirects.
func Validate(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only public http or https addresses are allowed")
	}
	port := parsed.Port()
	if port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return errors.New("invalid request port")
		}
	}
	if address := net.ParseIP(parsed.Hostname()); !AllowPrivate.Load() && address != nil && !PublicAddress(address) {
		return errors.New("the host is a forbidden address")
	}
	return nil
}
