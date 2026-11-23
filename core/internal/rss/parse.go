// Package rss fetches feeds over HTTP and parses them into store values. It
// reads RSS 0.9x, 1.0 and 2.0, Atom and JSON Feed.
package rss

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"

	"github.com/sdttttt/ressu/core/internal/sanitize"
	"github.com/sdttttt/ressu/core/internal/textmetrics"
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
	Enclosures  []Enclosure
}

// Enclosure is a media attachment of an item, such as a podcast episode.
type Enclosure struct {
	URL    string
	Type   string
	Length int64 // declared size in bytes, zero when the feed omits or garbles it
}

// Feed is a parsed feed.
type Feed struct {
	Meta  Meta
	Items []Item
}

// SummaryLimit bounds the plain text summary of an item.
const SummaryLimit = 280

// Parse reads the text of a feed. now dates the items that carry no date.
// strip enables tracking-parameter removal from the links of item bodies.
func Parse(data []byte, feedURL string, now time.Time, strip bool) (*Feed, error) {
	parsed, err := gofeed.NewParser().Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return convertFeed(parsed, feedURL, now, strip), nil
}

// maxEnclosures bounds the attachments kept for one item.
const maxEnclosures = 8

// enclosuresOf converts the attachments gofeed parsed, keeping only
// resolvable HTTP addresses and a bounded count per item.
func enclosuresOf(feedURL string, it *gofeed.Item) []Enclosure {
	if len(it.Enclosures) == 0 {
		return nil
	}
	enclosures := make([]Enclosure, 0, min(len(it.Enclosures), maxEnclosures))
	for _, e := range it.Enclosures {
		if e == nil || len(enclosures) >= maxEnclosures {
			continue
		}
		address := resolve(feedURL, e.URL)
		if address == "" {
			continue
		}
		enclosures = append(enclosures, Enclosure{
			URL:    address,
			Type:   truncate(strings.TrimSpace(e.Type), 64),
			Length: enclosureLength(e.Length),
		})
	}
	return enclosures
}

// enclosureLength parses the declared size, treating missing or malformed
// values as zero.
func enclosureLength(raw string) int64 {
	length, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || length < 0 {
		return 0
	}
	return length
}
func convertFeed(parsed *gofeed.Feed, feedURL string, now time.Time, strip bool) *Feed {
	feed := &Feed{
		Meta: Meta{
			Title:       strings.TrimSpace(parsed.Title),
			SiteURL:     resolve(feedURL, parsed.Link),
			Description: plainText(parsed.Description),
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
			Summary:     plainText(descriptionOf(it)),
			Content:     content,
			PublishedAt: published,
			Enclosures:  enclosuresOf(feedURL, it),
		}
		item.Content = sanitize.HTML(item.Content, item.URL, strip)
		item.WordCount = textmetrics.Count(item.Content)
		if item.WordCount == 0 {
			item.WordCount = textmetrics.Count(item.Summary)
		}
		if item.GUID == "" {
			// Items without a GUID, link, or date would otherwise collide on
			// the shared fetch timestamp; the content hash keeps same-batch
			// duplicates apart and stays stable across refreshes.
			sum := fnv.New32a()
			_, _ = io.WriteString(sum, it.Title)
			_, _ = io.WriteString(sum, "\x00")
			_, _ = io.WriteString(sum, it.Description)
			_, _ = io.WriteString(sum, "\x00")
			_, _ = io.WriteString(sum, it.Content)
			item.GUID = item.Title + "|" + item.PublishedAt.Format(time.RFC3339Nano) +
				fmt.Sprintf("|%08x", sum.Sum32())
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
// they are. References with a non-HTTP scheme are dropped: they would
// otherwise become the base that relative links in article bodies resolve
// against.
func resolve(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if r.Scheme != "" && r.Scheme != "http" && r.Scheme != "https" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	return b.ResolveReference(r).String()
}

// plainText turns markup into the text it shows, collapsed to one line and cut
// to SummaryLimit runes.
func plainText(markup string) string {
	if markup == "" {
		return ""
	}
	if !strings.ContainsAny(markup, "<&") {
		return truncate(collapse(markup), SummaryLimit)
	}

	z := html.NewTokenizer(strings.NewReader(markup))
	var b strings.Builder
	skipped := 0
	for {
		switch z.Next() {
		case html.ErrorToken:
			return truncate(collapse(b.String()), SummaryLimit)
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
