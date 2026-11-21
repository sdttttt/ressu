// Package extract pulls the readable body of an article page on demand.
// The result is cleaned by the same sanitize pipeline as feed content and
// is never stored.
package extract

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"

	"github.com/sdttttt/ressu/core/internal/limitreader"
	"github.com/sdttttt/ressu/core/internal/safedial"
	"github.com/sdttttt/ressu/core/internal/sanitize"
)

// MaxBody bounds the size of an article page.
const MaxBody = 8 << 20

// Timeout bounds fetching and extracting one article.
const Timeout = 15 * time.Second

// userAgent identifies the app to article servers.
const userAgent = "Ressu/0.1 (+https://github.com/sdttttt/ressu)"

// ErrExtraction reports a page whose readable content could not be found.
var ErrExtraction = errors.New("could not extract the article")

// client makes the requests. It refuses private-network addresses,
// including across redirects.
var client = &http.Client{
	Transport:     safedial.Transport(),
	CheckRedirect: safedial.CheckRedirect,
	Timeout:       Timeout,
}

// Article fetches raw and returns the sanitized readable body of the page.
// The address must be an absolute HTTP or HTTPS URL.
func Article(ctx context.Context, raw string) (string, error) {
	if err := safedial.Validate(raw); err != nil {
		return "", err
	}
	pageURL, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the server answered %s", res.Status)
	}

	body := limitreader.New(res.Body, MaxBody)
	article, err := readability.FromReader(body, pageURL)
	if bodyErr := body.Err(); bodyErr != nil {
		return "", bodyErr
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtraction, err)
	}
	var readable strings.Builder
	if err := article.RenderHTML(&readable); err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtraction, err)
	}
	if strings.TrimSpace(readable.String()) == "" {
		return "", ErrExtraction
	}
	return sanitize.HTML(readable.String(), raw), nil
}
