package rss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/sdttttt/ressu/pkg/limitreader"
)

// MaxBody bounds the size of a feed response.
const MaxBody = 16 << 20

// Timeout bounds a request and the reading of its body.
const Timeout = 15 * time.Second

// UserAgent identifies the app to feed servers.
const UserAgent = "Ressu/0.1 (+https://github.com/sdttttt/ressu)"

// accept lists the types feeds are served as, and falls back to anything.
const accept = "application/rss+xml, application/atom+xml, application/feed+json, " +
	"application/xml;q=0.9, text/xml;q=0.9, */*;q=0.8"

// ErrNotModified reports a 304: the feed did not change since the last fetch.
var ErrNotModified = errors.New("feed not modified")

// ErrInvalidURL reports an address that is not an http or https URL.
var ErrInvalidURL = errors.New("enter an http or https address")

// Result is a fetched and parsed feed with the validators of its response.
type Result struct {
	Feed         *Feed
	ETag         string
	LastModified string
}

// Client makes the requests.
var Client = &http.Client{Timeout: Timeout}

// ValidateURL checks that raw is an absolute http or https URL.
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrInvalidURL
	}
	return nil
}

// Fetch downloads and parses a feed. It sends etag and lastModified as
// validators and returns ErrNotModified when the server answers 304.
func Fetch(ctx context.Context, feedURL, etag, lastModified string) (*Result, error) {
	return fetch(ctx, feedURL, etag, lastModified, MaxBody)
}

// fetch implements Fetch with an explicit body limit.
func fetch(ctx context.Context, feedURL, etag, lastModified string, maxBody int64) (*Result, error) {
	if err := ValidateURL(feedURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", accept)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified {
		return nil, ErrNotModified
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the server answered %s", res.Status)
	}

	body := limitreader.New(res.Body, maxBody)
	parsed, err := gofeed.NewParser().Parse(body)
	if err := body.Err(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("this address does not serve a feed: %w", err)
	}

	// XML parsers can stop at the root's closing tag, so consume the rest to
	// enforce the limit and surface trailing read errors.
	if _, err := io.Copy(io.Discard, body); err != nil {
		return nil, err
	}

	return &Result{
		Feed:         convertFeed(parsed, feedURL, time.Now()),
		ETag:         res.Header.Get("ETag"),
		LastModified: res.Header.Get("Last-Modified"),
	}, nil
}
