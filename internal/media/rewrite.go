package media

import (
	"log/slog"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/sdttttt/ressu/internal/logging"
)

func cacheLocalURL(c *Cache, raw string) (string, bool) {
	source, ok := normalizeSource(raw)
	if !ok {
		return "", false
	}
	key := keyFor(source)
	if !cachePrepare(c, key, source) {
		return "", false
	}
	return localURL(key), true
}

func cacheRewriteHTML(c *Cache, markup, baseURL string) string {
	if markup == "" {
		return ""
	}
	document, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		return ""
	}
	body := bodyOf(document)
	if body == nil {
		slog.Warn("article image rewrite failed",
			"operation", "media.rewriteHTML", "reason", "no document body")
		return ""
	}
	stats := rewriteStats{}
	rewriteImages(body, baseURL, c, &stats)

	var output strings.Builder
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&output, child); err != nil {
			slog.Warn("article image rewrite failed",
				"operation", "media.rewriteHTML", "reason", "render failed", "error", err)
			return ""
		}
	}
	logging.Debug("article images rewritten",
		"operation", "media.rewriteHTML",
		"mappedImages", stats.mapped,
		"unmappedImages", stats.unmapped,
	)
	return output.String()
}

type rewriteStats struct {
	mapped   int
	unmapped int
}

// rewriteImages walks one cleaned article fragment and maps every img source.
func rewriteImages(node *html.Node, baseURL string, cache *Cache, stats *rewriteStats) {
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, "img") {
		if src := attribute(node, "src"); src != nil {
			if scheme, _, err := splitScheme(src.Val); err == nil &&
				(scheme == "data" || scheme == Scheme) {
				return
			}
			source := absolute(baseURL, src.Val)
			if local, ok := cache.LocalURL(source); ok {
				src.Val = local
				stats.mapped++
			} else {
				removeAttribute(node, "src")
				stats.unmapped++
			}
		}
		removeAttribute(node, "srcset")
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		rewriteImages(child, baseURL, cache, stats)
	}
}

// bodyOf finds the body element created while parsing a fragment.
func bodyOf(document *html.Node) *html.Node {
	if document.Type == html.ElementNode && document.Data == "body" {
		return document
	}
	for child := document.FirstChild; child != nil; child = child.NextSibling {
		if body := bodyOf(child); body != nil {
			return body
		}
	}
	return nil
}

// absolute resolves a reference against the article address.
func absolute(baseURL string, reference string) string {
	if baseURL == "" || reference == "" {
		return reference
	}
	base, err := url.Parse(baseURL)
	if err != nil || !base.IsAbs() {
		return reference
	}
	parsed, err := url.Parse(reference)
	if err != nil {
		return reference
	}
	return base.ResolveReference(parsed).String()
}

// splitScheme reports the lowercase scheme of an already-sanitized URL.
func splitScheme(raw string) (string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	return strings.ToLower(parsed.Scheme), raw, nil
}

// attribute returns a mutable attribute by its lowercase name.
func attribute(node *html.Node, name string) *html.Attribute {
	for index := range node.Attr {
		if strings.EqualFold(node.Attr[index].Key, name) {
			return &node.Attr[index]
		}
	}
	return nil
}

// removeAttribute removes an attribute by its lowercase name.
func removeAttribute(node *html.Node, name string) {
	position := -1
	for index := range node.Attr {
		if strings.EqualFold(node.Attr[index].Key, name) {
			position = index
			break
		}
	}
	if position >= 0 {
		node.Attr = append(node.Attr[:position], node.Attr[position+1:]...)
	}
}
