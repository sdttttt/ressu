// Package store keeps feeds, posts and preferences in a bbolt database under
// the application's user data directory.
package store

import (
	"math"
	"strings"
	"time"
	"unicode"
)

// Feed is a subscription and the metadata of its channel.
type Feed struct {
	ID              int64      `json:"id"`
	URL             string     `json:"url"`
	Title           string     `json:"title"`
	CustomTitle     bool       `json:"customTitle"`
	RemoteResources string     `json:"remoteResources"` // empty follows the global preference
	SiteURL         string     `json:"siteUrl"`
	Description     string     `json:"description"`
	IconURL         string     `json:"iconUrl"`
	Language        string     `json:"language"`
	ETag            string     `json:"etag"`
	LastModified    string     `json:"lastModified"`
	LastFetchedAt   *time.Time `json:"lastFetchedAt"`
	LastError       string     `json:"lastError"`
	// Unread is derived from the stored posts: recomputed for every feed when
	// the database opens, then maintained in memory. The persisted value is a
	// cache, never the source of truth.
	Unread    int       `json:"unread"`
	CreatedAt time.Time `json:"createdAt"`
}

// IconMeta describes a cached feed icon.
type IconMeta struct {
	SourceURL   string    `json:"sourceUrl"`
	ContentType string    `json:"contentType"`
	FetchedAt   time.Time `json:"fetchedAt"`
}

// Icon is the cached image of a feed, with the metadata needed to reuse it.
type Icon struct {
	Meta IconMeta `json:"meta"`
	Data []byte   `json:"data"`
}

// Enclosure is a media attachment of a post, such as a podcast episode.
type Enclosure struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Length int64  `json:"length"` // declared size in bytes, zero when unknown
}

// Post is an item of a feed, with the HTML of its body.
type Post struct {
	FeedID      int64       `json:"feedId"`
	GUID        string      `json:"guid"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	Author      string      `json:"author"`
	Summary     string      `json:"summary"`
	Content     string      `json:"content"`
	WordCount   int         `json:"wordCount"`
	PublishedAt time.Time   `json:"publishedAt"`
	FetchedAt   time.Time   `json:"fetchedAt"`
	Read        bool        `json:"read"`
	Starred     bool        `json:"starred"`
	Enclosures  []Enclosure `json:"enclosures"`
}

// ToSummary returns the post without its body, as lists show it.
func (p Post) ToSummary() PostSummary {
	return PostSummary{
		FeedID:      p.FeedID,
		GUID:        p.GUID,
		Title:       p.Title,
		URL:         p.URL,
		Author:      p.Author,
		Summary:     p.Summary,
		WordCount:   p.WordCount,
		PublishedAt: p.PublishedAt,
		FetchedAt:   p.FetchedAt,
		Read:        p.Read,
		Starred:     p.Starred,
	}
}

// PostSummary is a post as lists show it, without its body.
type PostSummary struct {
	FeedID      int64     `json:"feedId"`
	GUID        string    `json:"guid"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	Summary     string    `json:"summary"`
	WordCount   int       `json:"wordCount"`
	PublishedAt time.Time `json:"publishedAt"`
	FetchedAt   time.Time `json:"fetchedAt"`
	Read        bool      `json:"read"`
	Starred     bool      `json:"starred"`
}

// PostSkim is a Post without its body, as list queries decode it. Skipping
// the content field keeps the JSON decoder from allocating every article's
// HTML on the list path.
type PostSkim struct {
	FeedID      int64     `json:"feedId"`
	GUID        string    `json:"guid"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	Summary     string    `json:"summary"`
	WordCount   int       `json:"wordCount"`
	PublishedAt time.Time `json:"publishedAt"`
	FetchedAt   time.Time `json:"fetchedAt"`
	Read        bool      `json:"read"`
	Starred     bool      `json:"starred"`
}

// ToSummary returns the post without its body, as lists show it.
func (p PostSkim) ToSummary() PostSummary {
	return PostSummary(p)
}

// Theme values of Prefs.Theme.
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSepia  = "sepia"
)

// FontSize values of Prefs.FontSize.
const (
	FontSizeSmall  = "small"
	FontSizeMedium = "medium"
	FontSizeLarge  = "large"
)

// Values of Prefs.OpenLinks.
const (
	OpenLinksBrowser = "browser"
	OpenLinksWindow  = "window"
)

// Values of Prefs.PostSort, the article-list sort field.
const (
	// PostSortPublished orders posts by publication time, the default.
	PostSortPublished = "published"

	// PostSortFetched orders posts by the time Ressu fetched them.
	PostSortFetched = "fetched"
)

// Values of Prefs.RemoteResources.
const (
	ResourceAccessGoProxy = "goProxy"
	ResourceAccessWebview = "webview"
)

// Values of Prefs.UIFontFamily and Prefs.ArticleFontFamily. Any other value
// is a custom CSS font-family list entered by the user.
const (
	FontFamilySystem = "system"
	FontFamilySerif  = "serif"
	FontFamilySans   = "sans"
	FontFamilyMono   = "mono"
)

const (
	// MinArticleFontSize bounds a custom article font size, in px.
	MinArticleFontSize = 12
	// MaxArticleFontSize bounds a custom article font size, in px.
	MaxArticleFontSize = 28

	// MinArticleLineHeight bounds the article line height.
	MinArticleLineHeight = 1.2
	// DefaultArticleLineHeight is the article line height of a new install.
	DefaultArticleLineHeight = 1.78
	// MaxArticleLineHeight bounds the article line height.
	MaxArticleLineHeight = 2.4

	// MaxFontFamilyLen bounds a custom font-family value, in runes.
	MaxFontFamilyLen = 160
)

// fontFamilyStacks maps the preset family keywords to CSS font-family
// stacks.
var fontFamilyStacks = map[string]string{
	FontFamilySystem: `ui-sans-serif, system-ui, -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif`,
	FontFamilySerif:  `Georgia, "Times New Roman", "Songti SC", "SimSun", serif`,
	FontFamilySans:   `system-ui, -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif`,
	FontFamilyMono:   `ui-monospace, "Cascadia Mono", Consolas, "Sarasa Mono SC", monospace`,
}

// ResolveFontFamily maps a font family preference to a CSS font-family
// value: preset keywords expand to their stacks, empty selects the default
// system stack, and anything else is a custom family list used as given.
func ResolveFontFamily(value string) string {
	if stack, ok := fontFamilyStacks[value]; ok {
		return stack
	}
	if value == "" {
		return fontFamilyStacks[FontFamilySystem]
	}
	return value
}

// DefaultPostsPerFeed bounds the posts kept for a feed.
const DefaultPostsPerFeed = 500

const (
	// MinImageCacheLimit bounds the article image cache from below, in MB.
	MinImageCacheLimit = 32

	// DefaultImageCacheLimit is the article image cache limit of a new
	// install, in MB.
	DefaultImageCacheLimit = 256

	// MaxImageCacheLimit bounds the article image cache from above, in MB.
	MaxImageCacheLimit = 512
)

// Prefs are the user's preferences, kept in the settings bucket.
type Prefs struct {
	Theme             string  `json:"theme"`
	FontSize          string  `json:"fontSize"`
	RefreshOnStart    bool    `json:"refreshOnStart"`
	MarkReadOnOpen    bool    `json:"markReadOnOpen"`
	RemoteResources   string  `json:"remoteResources"`
	OpenLinks         string  `json:"openLinks"`
	PostSort          string  `json:"postSort"`
	StripTrackers     bool    `json:"stripTrackers"` // remove known tracking parameters from outbound links
	PostsPerFeed      int     `json:"postsPerFeed"`
	ImageCacheLimit   int     `json:"imageCacheLimit"` // maximum image cache size, in MB
	PaneSizes         []int   `json:"paneSizes"`       // widths of the sidebar and the list, in percent
	SidebarCollapsed  bool    `json:"sidebarCollapsed"`
	ReduceMotion      bool    `json:"reduceMotion"` // cut article swaps and native slides short
	UIFontFamily      string  `json:"uiFontFamily"` // preset keyword or custom CSS font-family; system is the default stack
	ArticleFontFamily string  `json:"articleFontFamily"`
	ArticleFontSize   int     `json:"articleFontSize"`   // custom article font size in px; zero follows the interface scale
	ArticleLineHeight float64 `json:"articleLineHeight"` // article line height, DefaultArticleLineHeight out of the box
}

// DefaultPrefs are the preferences of a new install.
func DefaultPrefs() Prefs {
	return Prefs{
		Theme:             ThemeSystem,
		FontSize:          FontSizeMedium,
		RefreshOnStart:    true,
		MarkReadOnOpen:    true,
		RemoteResources:   ResourceAccessWebview,
		OpenLinks:         OpenLinksBrowser,
		PostSort:          PostSortPublished,
		StripTrackers:     true,
		PostsPerFeed:      DefaultPostsPerFeed,
		ImageCacheLimit:   DefaultImageCacheLimit,
		PaneSizes:         []int{18, 30},
		SidebarCollapsed:  false,
		ReduceMotion:      false,
		UIFontFamily:      FontFamilySystem,
		ArticleFontFamily: FontFamilySystem,
		ArticleFontSize:   0,
		ArticleLineHeight: DefaultArticleLineHeight,
	}
}

// Normalize replaces the values that are missing or out of range.
func (p *Prefs) Normalize() {
	switch p.Theme {
	case ThemeSystem, ThemeLight, ThemeDark, ThemeSepia:
	default:
		p.Theme = ThemeSystem
	}
	switch p.FontSize {
	case FontSizeSmall, FontSizeMedium, FontSizeLarge:
	default:
		p.FontSize = FontSizeMedium
	}
	if p.PostSort != PostSortFetched {
		p.PostSort = PostSortPublished
	}
	if p.OpenLinks != OpenLinksWindow {
		p.OpenLinks = OpenLinksBrowser
	}
	if p.PostsPerFeed < 50 || p.PostsPerFeed > 2000 {
		p.PostsPerFeed = DefaultPostsPerFeed
	}
	if p.ImageCacheLimit < MinImageCacheLimit || p.ImageCacheLimit > MaxImageCacheLimit {
		p.ImageCacheLimit = DefaultImageCacheLimit
	}
	if p.RemoteResources != ResourceAccessWebview {
		p.RemoteResources = ResourceAccessGoProxy
	}
	if len(p.PaneSizes) != 2 || p.PaneSizes[0] < 5 || p.PaneSizes[0] > 90 ||
		p.PaneSizes[1] < 5 || p.PaneSizes[1] > 90 {
		p.PaneSizes = []int{18, 30}
	}
	p.UIFontFamily = normalizeFontFamily(p.UIFontFamily)
	p.ArticleFontFamily = normalizeFontFamily(p.ArticleFontFamily)
	if p.ArticleFontSize != 0 &&
		(p.ArticleFontSize < MinArticleFontSize || p.ArticleFontSize > MaxArticleFontSize) {
		p.ArticleFontSize = 0
	}
	if math.IsNaN(p.ArticleLineHeight) || p.ArticleLineHeight < MinArticleLineHeight ||
		p.ArticleLineHeight > MaxArticleLineHeight {
		p.ArticleLineHeight = DefaultArticleLineHeight
	}
}

// normalizeFontFamily keeps preset keywords (empty values from older
// databases migrate to the system default) and strips characters from
// custom values that could break out of a CSS font-family declaration.
func normalizeFontFamily(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "", FontFamilySystem, FontFamilySerif, FontFamilySans, FontFamilyMono:
		if value == "" {
			return FontFamilySystem
		}
		return value
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r == ';' || r == '{' || r == '}' || r == '<' || r == '>' || unicode.IsControl(r):
			continue
		case b.Len() >= MaxFontFamilyLen:
			continue
		default:
			b.WriteRune(r)
		}
	}
	cleaned := strings.Join(strings.Fields(b.String()), " ")
	if cleaned == "" {
		return FontFamilySystem
	}
	return cleaned
}
