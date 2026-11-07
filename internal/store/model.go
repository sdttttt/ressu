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

// Post is an item of a feed, with the HTML of its body.
type Post struct {
	FeedID      int64     `json:"feedId"`
	GUID        string    `json:"guid"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
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
		PublishedAt: p.PublishedAt,
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
	PublishedAt time.Time `json:"publishedAt"`
	Read        bool      `json:"read"`
	Starred     bool      `json:"starred"`
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

// DefaultPostsPerFeed bounds the posts kept for a feed.
const DefaultPostsPerFeed = 500

// Prefs are the user's preferences, kept in the settings bucket.
type Prefs struct {
	Theme          string `json:"theme"`
	FontSize       string `json:"fontSize"`
	RefreshOnStart bool   `json:"refreshOnStart"`
	MarkReadOnOpen bool   `json:"markReadOnOpen"`
	OpenLinks      string `json:"openLinks"`
	PostsPerFeed   int    `json:"postsPerFeed"`
	PaneSizes      []int  `json:"paneSizes"` // widths of the sidebar and the list, in percent
}

// DefaultPrefs are the preferences of a new install.
func DefaultPrefs() Prefs {
	return Prefs{
		Theme:          ThemeSystem,
		FontSize:       FontSizeMedium,
		RefreshOnStart: true,
		MarkReadOnOpen: true,
		OpenLinks:      OpenLinksBrowser,
		PostsPerFeed:   DefaultPostsPerFeed,
		PaneSizes:      []int{18, 30},
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
	if p.OpenLinks != OpenLinksWindow {
		p.OpenLinks = OpenLinksBrowser
	}
	if p.PostsPerFeed < 50 || p.PostsPerFeed > 2000 {
		p.PostsPerFeed = DefaultPostsPerFeed
	}
	if len(p.PaneSizes) != 2 || p.PaneSizes[0] < 5 || p.PaneSizes[0] > 90 ||
		p.PaneSizes[1] < 5 || p.PaneSizes[1] > 90 {
		p.PaneSizes = []int{18, 30}
	}
}
