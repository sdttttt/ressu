package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// outline is one <outline> element of an OPML document. Nested outlines
// (category folders) are walked recursively.
type outline struct {
	Type   string    `xml:"type,attr"`
	Text   string    `xml:"text,attr"`
	XMLURL string    `xml:"xmlUrl,attr"`
	Nested []outline `xml:"outline"`
}

// opml is the root of an OPML subscription list.
type opml struct {
	Body struct {
		Outlines []outline `xml:"outline"`
	} `xml:"body"`
}

// ParseOPML reads an OPML subscription list and returns the feed addresses
// it contains. Outlines without an xmlUrl attribute (category folders) are
// skipped.
func ParseOPML(data []byte) ([]string, error) {
	var doc opml
	if err := xml.NewDecoder(bytes.NewReader(data)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing OPML: %w", err)
	}

	var urls []string
	collect(doc.Body.Outlines, &urls)
	return urls, nil
}

// collect walks outlines recursively, gathering every xmlUrl.
func collect(outlines []outline, urls *[]string) {
	for _, o := range outlines {
		if o.XMLURL != "" {
			*urls = append(*urls, o.XMLURL)
		}
		collect(o.Nested, urls)
	}
}

// maxOPMLBytes bounds how large an OPML file can be.
const maxOPMLBytes = MaxBody

// ReadOPML opens and parses an OPML file, enforcing the size limit.
func ReadOPML(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxOPMLBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading OPML: %w", err)
	}
	if len(data) > maxOPMLBytes {
		return nil, fmt.Errorf("OPML file is larger than %d bytes", maxOPMLBytes)
	}
	return ParseOPML(data)
}

// OPMLFeed describes one subscription in an exported OPML document.
type OPMLFeed struct {
	Title       string
	URL         string
	SiteURL     string
	Description string
	Language    string
}

// exportOutline is one subscription outline in an OPML export.
type exportOutline struct {
	Type        string `xml:"type,attr"`
	Text        string `xml:"text,attr"`
	Title       string `xml:"title,attr,omitempty"`
	XMLURL      string `xml:"xmlUrl,attr"`
	HTMLURL     string `xml:"htmlUrl,attr,omitempty"`
	Description string `xml:"description,attr,omitempty"`
	Language    string `xml:"language,attr,omitempty"`
}

// exportOPML is the root of an exported OPML document.
type exportOPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    struct {
		Title string `xml:"title"`
	} `xml:"head"`
	Body struct {
		Outlines []exportOutline `xml:"outline"`
	} `xml:"body"`
}

// ExportOPML encodes subscriptions as an OPML 2.0 document. Subscription
// order is preserved, and a feed without a title uses its address as text.
func ExportOPML(feeds []OPMLFeed) ([]byte, error) {
	doc := exportOPML{Version: "2.0"}
	doc.Head.Title = "Ressu subscriptions"
	doc.Body.Outlines = make([]exportOutline, 0, len(feeds))
	for _, feed := range feeds {
		title := strings.TrimSpace(feed.Title)
		feedURL := strings.TrimSpace(feed.URL)
		if title == "" {
			title = feedURL
		}
		doc.Body.Outlines = append(doc.Body.Outlines, exportOutline{
			Type:        "rss",
			Text:        title,
			Title:       title,
			XMLURL:      feedURL,
			HTMLURL:     strings.TrimSpace(feed.SiteURL),
			Description: feed.Description,
			Language:    feed.Language,
		})
	}

	var data bytes.Buffer
	data.WriteString(xml.Header)
	encoder := xml.NewEncoder(&data)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("encoding OPML: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("encoding OPML: %w", err)
	}
	return data.Bytes(), nil
}
