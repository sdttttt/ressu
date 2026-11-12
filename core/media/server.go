package media

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sdttttt/ressu/core/logging"
)

const maxFailures = 10000

var loggedURLPattern = regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://)[^\s"'<>]+`)

func cacheServeHTTP(c *Cache, writer http.ResponseWriter, request *http.Request) {
	started := time.Now()
	status, cause := c.serveHTTP(writer, request, started)
	logging.Debug("media image request",
		"operation", "media.serveHTTP",
		"cacheKey", strings.TrimPrefix(request.URL.Path, "/i/"),
		"status", status,
		"elapsedMs", time.Since(started).Milliseconds(),
	)
	if status == http.StatusServiceUnavailable && request.Context().Err() == nil &&
		c.reportFailure(strings.TrimPrefix(request.URL.Path, "/i/"), started) {
		args := []any{
			"operation", "media.serveHTTP",
			"cacheKey", strings.TrimPrefix(request.URL.Path, "/i/"),
			"elapsedMs", time.Since(started).Milliseconds(),
		}
		if cause != nil {
			args = append(args, "error", diagnosticError(cause))
		}
		slog.Warn("media image unavailable", args...)
	}
}

// cacheServeRequest answers one media request and reports the response status.
func cacheServeRequest(
	c *Cache,
	writer http.ResponseWriter,
	request *http.Request,
	started time.Time,
) (int, error) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead ||
		!strings.EqualFold(request.URL.Host, "localhost") {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol method or host")
	}

	prefix := "/i/"
	if !strings.HasPrefix(request.URL.Path, prefix) {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol path")
	}
	key := strings.TrimPrefix(request.URL.Path, prefix)
	if !validKey(key) {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol key")
	}

	entry, err := c.get(request.Context(), key)
	if err != nil {
		if entry.Size > 0 {
			if request.Context().Err() == nil && c.reportFailure(key, started) {
				slog.Warn("stale image served after refresh failure",
					"operation", "media.serveHTTP",
					"cacheKey", key,
					"error", diagnosticError(err),
				)
			}
		} else {
			http.Error(writer, "image unavailable", http.StatusServiceUnavailable)
			return http.StatusServiceUnavailable, err
		}
	}

	file, err := os.Open(c.dataFile(key))
	if err != nil {
		http.Error(writer, "image unavailable", http.StatusServiceUnavailable)
		return http.StatusServiceUnavailable, fmt.Errorf("could not open image bytes: %w", err)
	}
	defer file.Close()

	writer.Header().Set("Content-Type", entry.ContentType)
	writer.Header().Set("Cache-Control", "private, max-age=300")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(writer, request, "", entry.FetchedAt, file)
	return http.StatusOK, nil
}

// cacheReportFailure keeps repeated failures for the same image from filling
// the local diagnostic log.
func cacheReportFailure(c *Cache, key string, failedAt time.Time) bool {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()

	if last, ok := c.failureTimes[key]; ok && failedAt.Sub(last) < time.Minute {
		return false
	}
	if len(c.failureTimes) >= maxFailures {
		return false
	}
	c.failureTimes[key] = failedAt
	return true
}

// diagnosticError removes image source URLs from an error before logging it.
func diagnosticError(err error) string {
	return loggedURLPattern.ReplaceAllString(err.Error(), "${1}[redacted]")
}
