// Package rss fetches feeds over HTTP and parses them into store values. It
// reads RSS 0.9x, 1.0 and 2.0, Atom and JSON Feed.
package rss

import (
	"bytes"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"
	"github.com/sdttttt/ressu/internal/sanitize"
	"github.com/sdttttt/ressu/internal/textmetrics"
	"golang.org/x/net/html"
)

// Meta is the channel metadata of a feed.
type Meta struct {
	Title       string
	SiteURL     string
	Description string
	IconURL     string
	Language    string
}

// Item is one entry of a feed.
type Item struct {
	GUID        string
	Title       string
	URL         string
	Author      string
	Summary     string
	Content     string
	WordCount   int
	PublishedAt time.Time
}

// Feed is a parsed feed.
type Feed struct {
	Meta  Meta
	Items []Item
}

// SummaryLimit bounds the plain text summary of an item.
const SummaryLimit = 280

// Parse reads the text of a feed. now dates the items that carry no date.
func Parse(data []byte, feedURL string, now time.Time) (*Feed, error) {
	parsed, err := gofeed.NewParser().Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return convertFeed(parsed, feedURL, now), nil
}

// convertFeed turns a gofeed feed into the shape stored by Ressu.
func convertFeed(parsed *gofeed.Feed, feedURL string, now time.Time) *Feed {
	feed := &Feed{
		Meta: Meta{
			Title:       strings.TrimSpace(parsed.Title),
			SiteURL:     resolve(feedURL, parsed.Link),
			Description: plainText(parsed.Description, SummaryLimit),
			Language:    parsed.Language,
		},
	}
	if parsed.Image != nil {
		feed.Meta.IconURL = resolve(feedURL, parsed.Image.URL)
	}

	for _, it := range parsed.Items {
		if it == nil {
			continue
		}
		published := now
		switch {
		case it.PublishedParsed != nil:
			published = *it.PublishedParsed
		case it.UpdatedParsed != nil:
			published = *it.UpdatedParsed
		}

		content := it.Content
		if content == "" {
			content = it.Description
		}

		link := resolve(feedURL, it.Link)
		guid := it.GUID
		if guid == "" {
			guid = link
		}

		item := Item{
			GUID:        guid,
			Title:       strings.TrimSpace(it.Title),
			URL:         link,
			Author:      authorOf(it),
			Summary:     plainText(descriptionOf(it), SummaryLimit),
			Content:     content,
			PublishedAt: published,
		}
		item.Content = sanitize.HTML(item.Content, item.URL)
		item.WordCount = textmetrics.Count(item.Content)
		if item.WordCount == 0 {
			item.WordCount = textmetrics.Count(item.Summary)
		}
		if item.GUID == "" {
			item.GUID = item.Title + "|" + item.PublishedAt.Format(time.RFC3339Nano)
		}
		if item.Title == "" {
			item.Title = item.GUID
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

func descriptionOf(it *gofeed.Item) string {
	if it.Description != "" {
		return it.Description
	}
	return it.Content
}

func authorOf(it *gofeed.Item) string {
	if it.Author != nil && it.Author.Name != "" {
		return it.Author.Name
	}
	for _, a := range it.Authors {
		if a != nil && a.Name != "" {
			return a.Name
		}
	}
	return ""
}

// resolve makes ref absolute against base, leaving absolute and empty ones as
// they are.
func resolve(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return b.ResolveReference(r).String()
}

// plainText turns markup into the text it shows, collapsed to one line and cut
// to limit runes.
func plainText(markup string, limit int) string {
	if markup == "" {
		return ""
	}
	if !strings.ContainsAny(markup, "<&") {
		return truncate(collapse(markup), limit)
	}

	z := html.NewTokenizer(strings.NewReader(markup))
	var b strings.Builder
	skipped := 0
	for {
		switch z.Next() {
		case html.ErrorToken:
			return truncate(collapse(b.String()), limit)
		case html.TextToken:
			if skipped == 0 {
				b.Write(z.Text())
			}
		case html.StartTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "script", "style":
				skipped++
			case "br", "p", "div", "li", "tr", "h1", "h2", "h3", "h4", "blockquote":
				b.WriteByte(' ')
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			if s := string(name); s == "script" || s == "style" {
				skipped = max(skipped-1, 0)
			}
		}
	}
}

func collapse(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}

func truncate(s string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:limit])) + "…"
}
