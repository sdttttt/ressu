package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
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
