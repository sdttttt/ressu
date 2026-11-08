package rss

import (
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

func TestParseOPML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want []string
	}{
		{
			name: "flat outline",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline type="rss" text="A" xmlUrl="https://a.example/feed.xml"/>
    <outline type="rss" text="B" xmlUrl="https://b.example/feed.xml"/>
  </body>
</opml>`,
			want: []string{"https://a.example/feed.xml", "https://b.example/feed.xml"},
		},
		{
			name: "nested outlines",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Tech">
      <outline type="rss" text="C" xmlUrl="https://c.example/feed.xml"/>
      <outline text="Deep">
        <outline type="rss" text="D" xmlUrl="https://d.example/feed.xml"/>
      </outline>
    </outline>
  </body>
</opml>`,
			want: []string{"https://c.example/feed.xml", "https://d.example/feed.xml"},
		},
		{
			name: "skips outlines without xmlUrl",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Just a folder"/>
    <outline type="rss" text="E" xmlUrl="https://e.example/feed.xml"/>
  </body>
</opml>`,
			want: []string{"https://e.example/feed.xml"},
		},
		{
			name: "empty body",
			xml:  `<opml version="1.0"><body></body></opml>`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOPML([]byte(tt.xml))
			if err != nil {
				t.Fatalf("ParseOPML() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOPML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOPMLRejectsInvalidXML(t *testing.T) {
	_, err := ParseOPML([]byte("not xml at all"))
	if err == nil {
		t.Error("ParseOPML(invalid) should fail")
	}
}

func TestReadOPML(t *testing.T) {
	urls, err := ReadOPML(strings.NewReader(
		`<opml version="1.0"><body><outline type="rss" xmlUrl="https://x.example/f.xml"/></body></opml>`,
	))
	if err != nil {
		t.Fatalf("ReadOPML() error = %v", err)
	}
	if len(urls) != 1 || urls[0] != "https://x.example/f.xml" {
		t.Errorf("ReadOPML() = %v, want one matching URL", urls)
	}
}

func TestExportOPML(t *testing.T) {
	data, err := ExportOPML([]OPMLFeed{{
		Title:       `A & "B"`,
		URL:         "https://a.example/feed.xml?a=1&b=2",
		SiteURL:     "https://a.example/",
		Description: "新闻 & 评论",
		Language:    "zh-CN",
	}})
	if err != nil {
		t.Fatalf("ExportOPML() error = %v", err)
	}

	var doc struct {
		Version string `xml:"version,attr"`
		Body    struct {
			Outlines []exportOutline `xml:"outline"`
		} `xml:"body"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal ExportOPML() = %v", err)
	}
	if doc.Version != "2.0" {
		t.Errorf("version = %q, want 2.0", doc.Version)
	}
	want := exportOutline{
		Type:        "rss",
		Text:        `A & "B"`,
		Title:       `A & "B"`,
		XMLURL:      "https://a.example/feed.xml?a=1&b=2",
		HTMLURL:     "https://a.example/",
		Description: "新闻 & 评论",
		Language:    "zh-CN",
	}
	if len(doc.Body.Outlines) != 1 || doc.Body.Outlines[0] != want {
		t.Errorf("outlines = %+v, want one outline %+v", doc.Body.Outlines, want)
	}
	urls, err := ParseOPML(data)
	if err != nil {
		t.Fatalf("ParseOPML(export) error = %v", err)
	}
	if len(urls) != 1 || urls[0] != want.XMLURL {
		t.Errorf("ParseOPML(export) = %v, want [%s]", urls, want.XMLURL)
	}
}

func TestExportOPMLUsesURLWithoutTitle(t *testing.T) {
	const feedURL = "https://a.example/feed.xml"
	data, err := ExportOPML([]OPMLFeed{{URL: feedURL}})
	if err != nil {
		t.Fatalf("ExportOPML() error = %v", err)
	}

	var doc struct {
		Body struct {
			Outlines []exportOutline `xml:"outline"`
		} `xml:"body"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal ExportOPML() = %v", err)
	}
	if len(doc.Body.Outlines) != 1 {
		t.Fatalf("outlines = %d, want 1", len(doc.Body.Outlines))
	}
	if got := doc.Body.Outlines[0]; got.Text != feedURL || got.Title != feedURL || got.XMLURL != feedURL {
		t.Errorf("outline = %+v, want all names to be %q", got, feedURL)
	}
}

func TestExportOPMLWithoutFeeds(t *testing.T) {
	data, err := ExportOPML(nil)
	if err != nil {
		t.Fatalf("ExportOPML() error = %v", err)
	}
	urls, err := ParseOPML(data)
	if err != nil {
		t.Fatalf("ParseOPML(empty export) error = %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("ParseOPML(empty export) = %v, want none", urls)
	}
}
