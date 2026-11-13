// Package store keeps feeds, posts and preferences in a bbolt database under
// the application's user data directory.
package store

import "time"

// Feed is a subscription and the metadata of its channel.
type Feed struct {
	ID            int64      `json:"id"`
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	SiteURL       string     `json:"siteUrl"`
	Description   string     `json:"description"`
	IconURL       string     `json:"iconUrl"`
	Language      string     `json:"language"`
	ETag          string     `json:"etag"`
	LastModified  string     `json:"lastModified"`
	LastFetchedAt *time.Time `json:"lastFetchedAt"`
	LastError     string     `json:"lastError"`
	Unread        int        `json:"unread"`
	CreatedAt     time.Time  `json:"createdAt"`
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

// Post is an item of a feed, with the HTML of its body.
type Post struct {
	FeedID      int64     `json:"feedId"`
	GUID        string    `json:"guid"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	WordCount   int       `json:"wordCount"`
	PublishedAt time.Time `json:"publishedAt"`
	FetchedAt   time.Time `json:"fetchedAt"`
	Read        bool      `json:"read"`
	Starred     bool      `json:"starred"`
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
	Theme            string `json:"theme"`
	FontSize         string `json:"fontSize"`
	RefreshOnStart   bool   `json:"refreshOnStart"`
	MarkReadOnOpen   bool   `json:"markReadOnOpen"`
	RemoteResources  string `json:"remoteResources"`
	OpenLinks        string `json:"openLinks"`
	PostSort         string `json:"postSort"`
	PostsPerFeed     int    `json:"postsPerFeed"`
	ImageCacheLimit  int    `json:"imageCacheLimit"` // maximum image cache size, in MB
	PaneSizes        []int  `json:"paneSizes"`       // widths of the sidebar and the list, in percent
	SidebarCollapsed bool   `json:"sidebarCollapsed"`
	ReduceMotion     bool   `json:"reduceMotion"` // cut article swaps and native slides short
}

// DefaultPrefs are the preferences of a new install.
func DefaultPrefs() Prefs {
	return Prefs{
		Theme:            ThemeSystem,
		FontSize:         FontSizeMedium,
		RefreshOnStart:   true,
		MarkReadOnOpen:   true,
		RemoteResources:  ResourceAccessWebview,
		OpenLinks:        OpenLinksBrowser,
		PostSort:         PostSortPublished,
		PostsPerFeed:     DefaultPostsPerFeed,
		ImageCacheLimit:  DefaultImageCacheLimit,
		PaneSizes:        []int{18, 30},
		SidebarCollapsed: false,
		ReduceMotion:     false,
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
}
