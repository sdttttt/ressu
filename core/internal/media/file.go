package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sdttttt/ressu/core/internal/safedial"
)

// MaxFileBytes bounds a single downloaded attachment, currently 256 MiB.
const MaxFileBytes = 256 << 20

// maxFileNameRunes bounds the attachment name suggested to the save dialog.
const maxFileNameRunes = 128

// fileClient downloads attachments without an overall timeout: large podcast
// episodes outlast any fixed deadline, and the caller's context still bounds
// the transfer.
var fileClient = &http.Client{
	CheckRedirect: safedial.CheckRedirect,
	Transport:     safedial.Transport(),
}

// ValidateSource reports whether raw is a public http or https address
// suitable for an attachment download.
func ValidateSource(raw string) error {
	if _, ok := normalizeSource(raw); !ok {
		return errors.New("unsupported attachment address")
	}
	return nil
}

// FileName suggests a safe file name for an attachment address: the last path
// segment without directories or query data, falling back to a generic name.
func FileName(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "attachment"
	}
	name := path.Base(parsed.Path)
	if name == "." || name == ".." || strings.ContainsAny(name, `/\\`) {
		return "attachment"
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "attachment"
	}
	runes := []rune(name)
	if len(runes) > maxFileNameRunes {
		name = string(runes[:maxFileNameRunes])
	}
	return name
}

// DownloadFile streams one remote attachment to destPath, which the native
// save dialog chose. The address must pass the public-URL policy; the
// transfer is bounded by MaxFileBytes, and a failed download removes the
// partial file.
func DownloadFile(ctx context.Context, rawURL, destPath string) error {
	source, ok := normalizeSource(rawURL)
	if !ok {
		return errors.New("unsupported attachment address")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := fileClient.Do(request)
	if err != nil {
		return fmt.Errorf("could not download the attachment: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("the attachment server answered %s", response.Status)
	}
	if response.ContentLength > MaxFileBytes {
		return errors.New("the attachment is larger than the local limit")
	}

	//nolint:gosec // G703: destPath comes from the user's own save dialog.
	file, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("could not create the attachment file: %w", err)
	}
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, MaxFileBytes+1))
	closeErr := file.Close()
	if copyErr == nil && closeErr != nil {
		copyErr = closeErr
	}
	if copyErr == nil && written > MaxFileBytes {
		copyErr = errors.New("the attachment is larger than the local limit")
	}
	if copyErr != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("could not save the attachment: %w", copyErr)
	}
	return nil
}
