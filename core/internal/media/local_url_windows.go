//go:build windows

package media

// localURL maps a cache key to the http URL WebView2 uses for the local
// protocol. It must match the filter MyGO registers on Windows.
func localURL(key string) string {
	return "http://" + Scheme + ".localhost/i/" + key
}
