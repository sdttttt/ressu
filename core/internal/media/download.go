package media

import (
	"bytes"
	"encoding/binary"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sdttttt/ressu/core/internal/safedial"
)

// userAgent identifies image requests without leaking article addresses in a
// browser-like Referer header.
const userAgent = "Ressu/0.1 (+https://github.com/sdttttt/ressu)"

// imageAccept lists raster image formats the target WebViews can render.
const imageAccept = "image/avif,image/png,image/jpeg,image/webp,image/gif,image/bmp,image/x-icon,*/*;q=0.8"

const avifContentType = "image/avif"

// allowedContentTypes covers formats recognized by http.DetectContentType.
// AVIF is handled separately by isAVIF.
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
		CheckRedirect: safedial.CheckRedirect,
		Transport:     safedial.Transport(),
		Timeout:       30 * time.Second,
	}
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
	if address := net.ParseIP(parsed.Hostname()); address != nil && !safedial.PublicAddress(address) {
		return "", false
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	source := parsed.String()
	if safedial.Validate(source) != nil {
		return "", false
	}
	return source, true
}

// imageType returns a supported bitmap type, preferring bytes over headers.
// AVIF needs a separate ISO BMFF signature check because http.DetectContentType
// does not recognize it.
func imageType(data []byte, headerType string) string {
	if isAVIF(data) {
		return avifContentType
	}
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

// isAVIF reports whether data starts with an AVIF file type box.
func isAVIF(data []byte) bool {
	if len(data) < 16 || !bytes.Equal(data[4:8], []byte("ftyp")) {
		return false
	}
	boxSize := binary.BigEndian.Uint32(data[:4])
	if boxSize < 16 || boxSize%4 != 0 || int64(boxSize) > int64(len(data)) {
		return false
	}
	for offset := 8; offset+4 <= int(boxSize); offset += 4 {
		if offset == 12 {
			continue
		}
		brand := string(data[offset : offset+4])
		if brand == "avif" || brand == "avis" {
			return true
		}
	}
	return false
}
