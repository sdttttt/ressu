// Package icon discovers and downloads a small bitmap icon for a feed.
package icon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/sdttttt/ressu/core/internal/rss"
)

// MaxBytes bounds the bytes kept for one downloaded icon.
const MaxBytes = 256 << 10

// maxHTMLBytes bounds the site page searched for favicon links.
const maxHTMLBytes = 1 << 20

// requestTimeout bounds one icon or site request.
const requestTimeout = 10 * time.Second

// ErrNoIcon reports that every icon candidate failed.
var ErrNoIcon = errors.New("no usable icon")

// Client makes icon requests.
var Client = &http.Client{Timeout: requestTimeout}

// Source describes where a feed icon may come from.
type Source struct {
	FeedURL string
	SiteURL string
	IconURL string
}

// Icon is a downloaded image and the address it came from.
type Icon struct {
	SourceURL   string
	ContentType string
	Data        []byte
}

// allowedContentTypes contains only bitmap formats that can safely render in
// an img element and are supported by the target WebViews.
var allowedContentTypes = map[string]bool{
	"image/gif":                true,
	"image/jpeg":               true,
	"image/png":                true,
	"image/webp":               true,
	"image/x-icon":             true,
	"image/vnd.microsoft.icon": true,
}

// Fetch discovers candidate addresses, downloads the first usable bitmap, and
// returns its source URL.
func Fetch(ctx context.Context, source Source) (Icon, error) {
	var lastErr error = ErrNoIcon
	for _, candidate := range candidates(ctx, source) {
		if err := ctx.Err(); err != nil {
			return Icon{}, err
		}

		image, err := download(ctx, candidate)
		if err == nil {
			return image, nil
		}
		lastErr = err
	}
	return Icon{}, fmt.Errorf("%w: %v", ErrNoIcon, lastErr)
}

// candidates returns the feed's explicit icon, favicon links on its site, and
// finally the site's /favicon.ico.
func candidates(ctx context.Context, source Source) []string {
	base := firstURL(source.SiteURL, source.FeedURL)
	var addresses []string
	if source.IconURL != "" {
		if validateURL(source.IconURL) == nil {
			addresses = append(addresses, strings.TrimSpace(source.IconURL))
		} else {
			addresses = append(addresses, resolve(base, source.IconURL))
		}
	}
	if base != "" {
		addresses = append(addresses, discover(ctx, base)...)
		addresses = append(addresses, resolve(base, "/favicon.ico"))
	}

	var unique []string
	for _, address := range addresses {
		if address == "" || slices.Contains(unique, address) {
			continue
		}
		unique = append(unique, address)
	}
	return unique
}

// discover downloads the site page and extracts icon link addresses.
func discover(parent context.Context, baseURL string) []string {
	body, err := getBody(parent, baseURL, "text/html,application/xhtml+xml,*/*;q=0.8", maxHTMLBytes)
	if err != nil {
		return nil
	}

	document, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	var addresses []string
	var visit func(node *html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "link") {
			var href string
			isIcon := false
			for _, attribute := range node.Attr {
				switch strings.ToLower(attribute.Key) {
				case "href":
					href = attribute.Val
				case "rel":
					tokens := strings.Fields(strings.ToLower(attribute.Val))
					isIcon = slices.Contains(tokens, "icon") ||
						slices.Contains(tokens, "apple-touch-icon") ||
						(slices.Contains(tokens, "shortcut") && slices.Contains(tokens, "icon"))
				}
			}
			if isIcon && href != "" {
				addresses = append(addresses, resolve(baseURL, href))
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return addresses
}

// download fetches one candidate and validates it as a bounded bitmap.
func download(ctx context.Context, address string) (Icon, error) {
	if err := validateURL(address); err != nil {
		return Icon{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Icon{}, err
	}
	request.Header.Set("User-Agent", rss.UserAgent)
	request.Header.Set("Accept", "image/png,image/jpeg,image/webp,image/gif,image/x-icon,*/*;q=0.8")

	response, err := Client.Do(request)
	if err != nil {
		return Icon{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Icon{}, fmt.Errorf("the server answered %s", response.Status)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, MaxBytes+1))
	if err != nil {
		return Icon{}, err
	}
	if len(data) > MaxBytes {
		return Icon{}, fmt.Errorf("icon is larger than %d bytes", MaxBytes)
	}

	contentType := imageType(data, response.Header.Get("Content-Type"))
	if contentType == "" {
		return Icon{}, errors.New("the icon is not a supported bitmap")
	}
	return Icon{SourceURL: address, ContentType: contentType, Data: data}, nil
}

// getBody fetches a bounded response body with the given accepted media types.
func getBody(parent context.Context, address, accept string, limit int64) ([]byte, error) {
	if err := validateURL(address); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(parent, requestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", rss.UserAgent)
	request.Header.Set("Accept", accept)

	response, err := Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the server answered %s", response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, limit+1))
}

// imageType returns a supported bitmap media type, preferring the type
// detected from the bytes over an untrusted response header.
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

// firstURL returns the first parseable absolute http or https address.
func firstURL(values ...string) string {
	for _, value := range values {
		if validateURL(value) == nil {
			return value
		}
	}
	return ""
}

// resolve makes ref absolute against base when base is valid.
func resolve(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || base == "" {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	resolved := baseURL.ResolveReference(refURL)
	if validateURL(resolved.String()) != nil {
		return ""
	}
	return resolved.String()
}

// validateURL rejects addresses that are not absolute http or https URLs.
func validateURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("icon address must be http or https")
	}
	return nil
}
