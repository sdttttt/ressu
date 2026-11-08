package media

import (
	"context"
	"errors"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// userAgent identifies image requests without leaking article addresses in a
// browser-like Referer header.
const userAgent = "Ressu/0.1 (+https://github.com/sdttttt/ressu)"

// imageAccept lists raster image formats the target WebViews can render.
const imageAccept = "image/png,image/jpeg,image/webp,image/gif,image/bmp,image/x-icon,*/*;q=0.8"

// allowedContentTypes is based on bytes detected below, not untrusted markup.
var allowedContentTypes = map[string]bool{
	"image/bmp":                true,
	"image/gif":                true,
	"image/jpeg":               true,
	"image/png":                true,
	"image/webp":               true,
	"image/x-icon":             true,
	"image/vnd.microsoft.icon": true,
}

// newClient makes a client with bounded redirects and private-network dialing
// disabled.
func newClient() *http.Client {
	return &http.Client{
		CheckRedirect: checkRedirect,
		Transport: &http.Transport{
			DialContext:           safeDial,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       time.Minute,
			MaxIdleConns:          16,
			MaxIdleConnsPerHost:   2,
		},
		Timeout: 30 * time.Second,
	}
}

// checkRedirect limits redirects and re-checks each target address.
func checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirect {
		return errors.New("too many image redirects")
	}
	if err := validateSource(request.URL.String()); err != nil {
		return err
	}
	if len(via) > 0 && via[0].URL.Scheme == "https" && request.URL.Scheme != "https" {
		return errors.New("image redirects may not downgrade to http")
	}
	return nil
}

// safeDial resolves the request host once, rejects unsafe addresses, and dials
// the validated IP so DNS cannot change between checking and connecting.
func safeDial(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, errors.New("only TCP image requests are allowed")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || port == "" {
		return nil, errors.New("invalid image host")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber != 80 && portNumber != 443 {
		return nil, errors.New("image ports are limited to http and https")
	}

	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	selected := make([]net.IP, 0, len(addresses))
	for _, candidate := range addresses {
		if !publicAddress(candidate.IP) {
			return nil, errors.New("the image host resolves to a forbidden address")
		}
		if network == "tcp" ||
			network == "tcp4" && candidate.IP.To4() != nil ||
			network == "tcp6" && candidate.IP.To4() == nil {
			selected = append(selected, candidate.IP)
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("the image host has no usable address")
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return dialer.DialContext(ctx, network, net.JoinHostPort(selected[0].String(), port))
}

// publicAddress rejects loopback, private, shared, link-local, multicast, and
// unspecified addresses.
func publicAddress(address net.IP) bool {
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

// normalizeSource accepts only ordinary public http and https image URLs.
func normalizeSource(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}
	port := parsed.Port()
	if port != "" && port != "80" && port != "443" {
		return "", false
	}
	if address := net.ParseIP(parsed.Hostname()); address != nil && !publicAddress(address) {
		return "", false
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	source := parsed.String()
	if validateSource(source) != nil {
		return "", false
	}
	return source, true
}

// validateSource is the protocol-level check applied after redirects too.
func validateSource(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only public http or https images can be cached")
	}
	port := parsed.Port()
	if port != "" && port != "80" && port != "443" {
		return errors.New("image ports are limited to http and https")
	}
	if address := net.ParseIP(parsed.Hostname()); address != nil && !publicAddress(address) {
		return errors.New("the image host is a forbidden address")
	}
	return nil
}

// imageType returns a supported bitmap type, preferring bytes over headers.
func imageType(data []byte, headerType string) string {
	detected := http.DetectContentType(data)
	if allowedContentTypes[detected] {
		return detected
	}
	parsed, _, err := mime.ParseMediaType(headerType)
	if err == nil && allowedContentTypes[parsed] {
		return parsed
	}
	return ""
}
