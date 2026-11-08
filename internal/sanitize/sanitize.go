// Package sanitize cleans untrusted HTML before it is stored for WebView
// rendering.
package sanitize

import (
	"net/url"
	"slices"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

// policy is safe for concurrent use after construction.
var policy = newPolicy()

// newPolicy builds the reading profile used by article bodies. bluemonday
// performs the security allowlisting; rendering-only attributes are added
// afterward by enhance.
func newPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	// Keep the markup produced by common feed generators, while leaving
	// forms, embedded documents, scripts, styles, SVG, and templates out.
	p.AllowElements(
		"address", "article", "aside", "audio", "bdi", "bdo", "big",
		"blink", "canvas", "center", "col", "colgroup", "data", "datalist",
		"decorator", "dialog", "dir", "element", "fieldset", "figure",
		"font", "footer", "header", "hgroup", "hr", "img", "kbd", "label",
		"legend", "main", "map", "mark", "marquee", "menu", "menuitem",
		"meter", "nav", "nobr", "optgroup", "option", "output", "picture",
		"progress", "q", "source", "span", "track", "video",
	)

	// Inert global attributes used by feed markup and accessibility metadata.
	p.AllowStyling()
	p.AllowDataAttributes()
	p.AllowAttrs(
		"align", "bgcolor", "border", "cellpadding", "cellspacing", "clear",
		"color", "cols", "colspan", "datetime", "height", "hidden", "ismap",
		"maxlength", "media", "min", "nowrap", "optimum", "part", "pattern",
		"placeholder", "rows", "rowspan", "scope", "size", "sizes", "span",
		"start", "tabindex", "translate", "valign", "value", "width", "wrap",
	).Globally()
	p.AllowAttrs(
		"aria-describedby", "aria-hidden", "aria-label", "aria-labelledby",
		"role",
	).Globally()
	p.AllowAttrs("for").OnElements("label", "output")

	// Links use the same address schemes DOMPurify considered safe.
	p.AllowURLSchemes(
		"ftp", "ftps", "http", "https", "mailto", "tel", "callto", "sms",
		"cid", "xmpp", "matrix",
	)
	p.AllowDataURIImages()
	p.AllowAttrs("download", "hreflang", "href", "rel", "target", "type").OnElements("a", "area")

	// Native media remains available, but source URLs still pass through the
	// URL policy above and srcset remains forbidden.
	p.AllowAttrs(
		"autoplay", "controls", "controlslist", "crossorigin", "loop", "muted",
		"playsinline", "preload",
	).OnElements("audio", "video")
	p.AllowAttrs("poster").OnElements("video")
	p.AllowAttrs("loading", "decoding").OnElements("img")
	p.AllowAttrs("src").OnElements("audio", "source", "track", "video")
	p.AllowAttrs("kind", "label", "srclang").OnElements("track")
	p.AllowAttrs("media", "type").OnElements("source")

	// bluemonday's default skips script, style, iframe, and object content.
	// These extra namespaces and inert roots also hide their descendants.
	p.SkipElementsContent(
		"annotation-xml", "foreignobject", "math", "plaintext", "svg",
		"template", "xmp",
	)

	return p
}

// HTML removes unsafe markup from an HTML fragment and resolves links and
// images against baseURL. Unsafe attributes are removed, and ordinary text
// remains intact.
func HTML(markup string, baseURL string) string {
	if markup == "" {
		return ""
	}

	clean := policy.Sanitize(markup)
	document, err := html.Parse(strings.NewReader(clean))
	if err != nil {
		return ""
	}
	body := bodyOf(document)
	if body == nil {
		return ""
	}

	var output strings.Builder
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		enhance(child, baseURL)
		if err := html.Render(&output, child); err != nil {
			return ""
		}
	}
	return output.String()
}

// bodyOf finds the body element created for an HTML fragment.
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

// enhance performs rendering adjustments after bluemonday sanitization:
// absolute URLs, safe external links, and lazy images.
func enhance(node *html.Node, baseURL string) {
	if node.Type == html.ElementNode {
		name := strings.ToLower(node.Data)
		if name == "a" {
			if href := attribute(node, "href"); href != nil {
				href.Val = absolute(baseURL, href.Val)
				setAttribute(node, "target", "_blank")
				setAttribute(node, "rel", "noopener noreferrer")
			}
		}
		if name == "img" {
			if src := attribute(node, "src"); src != nil {
				src.Val = absolute(baseURL, src.Val)
			}
			setAttribute(node, "loading", "lazy")
			setAttribute(node, "decoding", "async")
			removeAttribute(node, "width")
			removeAttribute(node, "height")
		}
		if name == "source" {
			if src := attribute(node, "src"); src != nil {
				src.Val = absolute(baseURL, src.Val)
				if parsed, err := url.Parse(src.Val); err != nil ||
					(parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https") {
					removeAttribute(node, "src")
				}
			}
		}
		// bluemonday can retain policy-related attributes in different orders
		// depending on whether they came from the feed or were required by
		// policy. Sorting makes repeated sanitization byte-for-byte stable.
		slices.SortFunc(node.Attr, func(first, second html.Attribute) int {
			return strings.Compare(strings.ToLower(first.Key), strings.ToLower(second.Key))
		})
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		enhance(child, baseURL)
	}
}

// absolute resolves reference against an absolute HTTP-like base URL.
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

// attribute returns a mutable attribute by its lowercase name.
func attribute(node *html.Node, name string) *html.Attribute {
	for i := range node.Attr {
		if strings.EqualFold(node.Attr[i].Key, name) {
			return &node.Attr[i]
		}
	}
	return nil
}

// setAttribute replaces or appends an attribute.
func setAttribute(node *html.Node, name string, value string) {
	if existing := attribute(node, name); existing != nil {
		existing.Val = value
		return
	}
	node.Attr = append(node.Attr, html.Attribute{Key: name, Val: value})
}

// removeAttribute removes an attribute by its lowercase name.
func removeAttribute(node *html.Node, name string) {
	node.Attr = slices.DeleteFunc(node.Attr, func(attr html.Attribute) bool {
		return strings.EqualFold(attr.Key, name)
	})
}
