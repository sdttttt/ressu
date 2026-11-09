package media

import (
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"strings"
)

// ErrInvalidImageSource reports that an address is not an image source this
// process may load.
var ErrInvalidImageSource = errors.New("invalid image source")

// ErrImageUnavailable reports that a mapped image has no usable bytes.
var ErrImageUnavailable = errors.New("image unavailable")

// IsAVIF reports whether data has an AVIF still-image or sequence signature.
func IsAVIF(data []byte) bool {
	return imageType(data, "") == avifContentType
}

// ContentType reports the supported bitmap type carried by image bytes.
func ContentType(data []byte) string {
	return imageType(data, "")
}

// localImageKey extracts the validated cache key from either local URL form.
// Accepting both forms keeps viewer input valid as an image crosses platform
// boundaries in tests or stored content.
func localImageKey(source string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(source))
	if err != nil || parsed.Opaque != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Hostname())
	allowedHost := scheme == Scheme && host == "localhost" ||
		scheme == "http" && host == Scheme+".localhost"
	if !allowedHost || !strings.HasPrefix(parsed.Path, "/i/") {
		return "", false
	}
	key := strings.TrimPrefix(parsed.Path, "/i/")
	return key, validKey(key)
}

// decodeDataURL decodes base64 data URLs within the image cache size limit.
func decodeDataURL(source string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(source))
	if err != nil || !strings.EqualFold(parsed.Scheme, "data") || parsed.Opaque == "" {
		return nil, ErrInvalidImageSource
	}

	metadata, encoded, found := strings.Cut(parsed.Opaque, ",")
	if !found || !strings.HasSuffix(strings.ToLower(metadata), ";base64") {
		return nil, ErrInvalidImageSource
	}
	if len(encoded) > base64.StdEncoding.EncodedLen(MaxBytes)+8 {
		return nil, errors.New("the image is larger than the local limit")
	}

	decoder := base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(encoded))
	data, err := io.ReadAll(io.LimitReader(decoder, MaxBytes+1))
	if err != nil {
		return nil, ErrInvalidImageSource
	}
	if len(data) > MaxBytes {
		return nil, errors.New("the image is larger than the local limit")
	}
	return data, nil
}
