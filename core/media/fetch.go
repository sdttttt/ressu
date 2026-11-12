package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxRedirect = 4

// cacheGet returns a fresh cached image, revalidates a stale one, downloads a
// prepared one, or reports why no usable bytes are available. A stale image is
// returned together with the refresh error.
func cacheGet(c *Cache, ctx context.Context, key string) (Entry, error) {
	return cacheGetOnce(c, ctx, key, 0)
}

func cacheGetOnce(c *Cache, ctx context.Context, key string, depth int) (Entry, error) {
	if depth > maxRedirect {
		return Entry{}, errors.New("too many image cache redirects")
	}

	c.mu.Lock()
	entry, ok := c.entries[key]
	if !ok {
		c.mu.Unlock()
		return Entry{}, errCacheMiss
	}
	if c.imageFresh(entry) {
		entry.LastAttempt = time.Now()
		c.entries[key] = entry
		c.mu.Unlock()
		return entry, nil
	}
	if imageNegative(entry, c.openedAt) && entry.Size == 0 {
		c.mu.Unlock()
		return Entry{}, errors.New("the image recently failed to download")
	}
	if existing, waiting := c.inFlight[key]; waiting {
		c.mu.Unlock()
		<-existing.done
		return c.getOnce(ctx, key, depth+1)
	}
	current := &fetchRequest{done: make(chan struct{})}
	c.inFlight[key] = current
	c.mu.Unlock()

	fetched, err := cacheFetch(c, ctx, entry)
	if err != nil && entry.Size > 0 {
		c.mu.Lock()
		delete(c.inFlight, key)
		close(current.done)
		c.mu.Unlock()
		entry.FailedAt = time.Now()
		entry.LastAttempt = entry.FailedAt
		return entry, err
	}
	c.mu.Lock()
	delete(c.inFlight, key)
	close(current.done)
	c.mu.Unlock()
	if err != nil {
		return Entry{}, err
	}
	return fetched, nil
}

// cacheImageFresh reports whether usable image bytes are still current enough.
func cacheImageFresh(entry Entry) bool {
	return entry.Size > 0 && entry.FetchedAt.Add(freshTTL).After(time.Now())
}

// imageNegative reports whether a recent failure from this process should not
// be retried. Failures from an earlier launch are retried once after restart.
func imageNegative(entry Entry, openedAt time.Time) bool {
	return entry.Size == 0 && !entry.FailedAt.IsZero() &&
		!entry.FailedAt.Before(openedAt) &&
		entry.FailedAt.Add(failureTTL).After(time.Now())
}

// cacheFetch downloads or revalidates one image and stores the accepted bytes.
func cacheFetch(c *Cache, ctx context.Context, old Entry) (Entry, error) {
	now := time.Now()
	entry := old
	entry.LastAttempt = now
	if old.Size == 0 {
		entry.CreatedAt = now
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, old.SourceURL, nil)
	if err != nil {
		c.fail(old, now)
		return Entry{}, err
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", imageAccept)
	if old.ETag != "" {
		request.Header.Set("If-None-Match", old.ETag)
	}
	if old.LastModified != "" {
		request.Header.Set("If-Modified-Since", old.LastModified)
	}

	response, err := c.client.Do(request)
	if err != nil {
		if request.Context().Err() != nil {
			return Entry{}, err
		}
		c.fail(old, now)
		return Entry{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotModified && old.Size > 0 {
		entry.ContentType = old.ContentType
		entry.FetchedAt = now
		entry.FailedAt = time.Time{}
		return entry, cachePut(c, old.Key, entry, nil)
	}
	if response.StatusCode != http.StatusOK {
		c.fail(old, now)
		return Entry{}, fmt.Errorf("the image server answered %s", response.Status)
	}
	if response.ContentLength > MaxBytes {
		c.fail(old, now)
		return Entry{}, errors.New("the image is larger than the local limit")
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, MaxBytes+1))
	if err != nil {
		if request.Context().Err() != nil {
			return Entry{}, err
		}
		c.fail(old, now)
		return Entry{}, err
	}
	if len(data) > MaxBytes {
		c.fail(old, now)
		return Entry{}, errors.New("the image is larger than the local limit")
	}

	contentType := imageType(data, response.Header.Get("Content-Type"))
	if contentType == "" {
		c.fail(old, now)
		return Entry{}, errors.New("the address does not serve a supported image")
	}

	entry.ContentType = contentType
	entry.ETag = response.Header.Get("ETag")
	entry.LastModified = response.Header.Get("Last-Modified")
	entry.FetchedAt = now
	entry.FailedAt = time.Time{}
	entry.Size = int64(len(data))
	return entry, cachePut(c, old.Key, entry, data)
}

// cacheFail records a short-lived negative result while retaining stale bytes.
func cacheFail(c *Cache, old Entry, failedAt time.Time) {
	entry := old
	entry.LastAttempt = failedAt
	entry.FailedAt = failedAt
	_ = cachePut(c, old.Key, entry, nil)
}
