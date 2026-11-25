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
// remains intact. When strip is true, known tracking parameters are also
// removed from the addresses of outbound links.
func HTML(markup string, baseURL string, strip bool) string {
	if markup == "" {
		return ""
	}
	if strings.TrimSpace(markup) == "" {
		return markup
	}

	clean := policy.SanitizeReader(strings.NewReader(markup))
	cleanLen := clean.Len()
	document, err := html.Parse(clean)
	if err != nil {
		return ""
	}
	body := bodyOf(document)
	if body == nil {
		return ""
	}

	var output strings.Builder
	output.Grow(cleanLen)
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		enhance(child, baseURL, strip)
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

// FirstImage returns the address of the first loadable image in sanitized
// markup, or "" when the markup has none. The markup must already have gone
// through HTML, so image addresses are absolute; data URIs are skipped
// because storing them outside the body would duplicate their bytes.
func FirstImage(markup string) string {
	if markup == "" {
		return ""
	}
	document, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		return ""
	}
	body := bodyOf(document)
	if body == nil {
		return ""
	}
	return firstImageIn(body)
}

// firstImageIn walks the tree in document order and returns the first img
// source the media cache may load.
func firstImageIn(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "img" {
		if src := attribute(node, "src"); src != nil {
			if parsed, err := url.Parse(src.Val); err == nil && parsed.Host != "" &&
				(parsed.Scheme == "http" || parsed.Scheme == "https") {
				return src.Val
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := firstImageIn(child); found != "" {
			return found
		}
	}
	return ""
}

// enhance performs rendering adjustments after bluemonday sanitization:
// absolute URLs, safe external links, lazy images, and optionally stripped
// tracking parameters on outbound links.
func enhance(node *html.Node, baseURL string, strip bool) {
	if node.Type == html.ElementNode {
		name := strings.ToLower(node.Data)
		if name == "a" {
			if href := attribute(node, "href"); href != nil {
				href.Val = absolute(baseURL, href.Val)
				if strip {
					href.Val = StripTrackingParams(href.Val, hostnameOf(baseURL))
				}
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
		enhance(child, baseURL, strip)
	}
}

// trackingPrefixes name the query parameter families stripped from outbound
// link addresses: UTM (Google Analytics) and Matomo campaign parameters.
var trackingPrefixes = []string{"utm_", "mtm_"}

// trackingParams names the exact query parameters stripped from outbound
// link addresses. The list follows Miniflux's urlcleaner, plus Alibaba's
// spm, and stays conservative: parameters that some sites use legitimately
// (for example ref and from) live in trackingParamsSelf instead.
var trackingParams = map[string]bool{
	// Facebook
	"fbclid":          true,
	"fb_action_ids":   true,
	"fb_action_types": true,
	"fb_ref":          true,
	"fb_source":       true,
	"fb_comment_id":   true,
	"_openstat":       true,

	// Google
	"gclid":         true,
	"dclid":         true,
	"gbraid":        true,
	"wbraid":        true,
	"gclsrc":        true,
	"srsltid":       true,
	"campaign_id":   true,
	"campaign_name": true,

	// Yandex, Twitter, Microsoft
	"yclid":   true,
	"ysclid":  true,
	"twclid":  true,
	"msclkid": true,

	// ESP and product analytics click identifiers
	"mc_cid":             true,
	"mc_eid":             true,
	"mc_tc":              true,
	"mkt_tok":            true,
	"wickedid":           true,
	"vero_id":            true,
	"vero_conv":          true,
	"hsa_cam":            true,
	"_hsenc":             true,
	"__hssc":             true,
	"__hstc":             true,
	"__hsfp":             true,
	"_hsmi":              true,
	"hsctatracking":      true,
	"rb_clickid":         true,
	"oly_anon_id":        true,
	"oly_enc_id":         true,
	"sc_cid":             true,
	"_bhlid":             true,
	"_branch_match_id":   true,
	"_branch_referrer":   true,
	"__readwiselocation": true,

	// Humble Bundle, ITM, Alibaba
	"hmb_campaign": true,
	"hmb_medium":   true,
	"hmb_source":   true,
	"itm_campaign": true,
	"itm_medium":   true,
	"itm_source":   true,
	"spm":          true,
}

// trackingParamsSelf names query parameters that are tracking only when
// their value points back at the originating site (Ghost-style self
// references on outbound links).
var trackingParamsSelf = map[string]bool{
	"ref":  true,
	"from": true,
}

// StripTrackingParams removes known tracking parameters from raw, an outbound
// link address. siteHost limits ref/from stripping to values that name the
// originating site; pass "" to skip that rule, for example when the link no
// longer carries its article context. Addresses without a host or a query
// are returned unchanged.
func StripTrackingParams(raw, siteHost string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.RawQuery == "" {
		return raw
	}
	query := parsed.Query()
	cleaned := false
	for name := range query {
		lower := strings.ToLower(name)
		if isTrackingParam(lower) {
			query.Del(name)
			cleaned = true
			continue
		}
		if !trackingParamsSelf[lower] || siteHost == "" {
			continue
		}
		for _, value := range query[name] {
			if strings.EqualFold(strings.TrimSpace(value), siteHost) {
				query.Del(name)
				cleaned = true
				break
			}
		}
	}
	if !cleaned {
		return raw
	}
	parsed.RawQuery = query.Encode()
	return strings.TrimSuffix(parsed.String(), "?")
}

func isTrackingParam(lower string) bool {
	for _, prefix := range trackingPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return trackingParams[lower]
}

// hostnameOf returns the host of an absolute HTTP-like base URL, or "" when
// the base cannot provide one.
func hostnameOf(baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil || !base.IsAbs() ||
		(base.Scheme != "http" && base.Scheme != "https") {
		return ""
	}
	return base.Hostname()
}

// absolute resolves reference against an absolute HTTP-like base URL. A base
// with any other scheme is ignored so it cannot be inherited by relative
// references.
func absolute(baseURL string, reference string) string {
	if baseURL == "" || reference == "" {
		return reference
	}
	base, err := url.Parse(baseURL)
	if err != nil || !base.IsAbs() || (base.Scheme != "http" && base.Scheme != "https") {
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
