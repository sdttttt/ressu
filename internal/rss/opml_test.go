package rss

import (
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
