//go:build !windows

package media

// localURL maps a cache key to the WebView's local protocol URL.
func localURL(key string) string {
	return Scheme + "://localhost/i/" + key
}
