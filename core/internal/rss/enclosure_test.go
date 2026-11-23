package rss

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParseKeepsEnclosures(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Podcast</title>
    <item>
      <title>Episode 1</title>
      <guid>episode-1</guid>
      <enclosure url="https://cdn.example.com/ep1.mp3" type="audio/mpeg" length="10485760"/>
      <enclosure url="/ep1-notes.pdf" type="application/pdf" length="2048"/>
      <enclosure url="file:///etc/passwd" type="text/plain" length="1"/>
      <enclosure url="https://cdn.example.com/bad.mp3" type="audio/mpeg" length="not-a-number"/>
    </item>
  </channel>
</rss>`
	feed, err := Parse([]byte(data), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(feed.Items))
	}

	enclosures := feed.Items[0].Enclosures
	if len(enclosures) != 3 {
		t.Fatalf("enclosures = %d, want 3 (the file:// one is dropped)", len(enclosures))
	}

	first := enclosures[0]
	if first.URL != "https://cdn.example.com/ep1.mp3" ||
		first.Type != "audio/mpeg" || first.Length != 10485760 {
		t.Errorf("first enclosure = %+v", first)
	}
	second := enclosures[1]
	if second.URL != "https://example.com/ep1-notes.pdf" {
		t.Errorf("relative enclosure resolved to %q", second.URL)
	}
	third := enclosures[2]
	if third.Length != 0 {
		t.Errorf("garbled length kept as %d, want 0", third.Length)
	}
}

func TestParseCapsEnclosureCount(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>f</title><item><title>i</title>`)
	for i := range 20 {
		b.WriteString(`<enclosure url="https://cdn.example.com/f` + strconv.Itoa(i) + `" type="audio/mpeg"/>`)
	}
	b.WriteString(`</item></channel></rss>`)
	feed, err := Parse([]byte(b.String()), "https://example.com/feed.xml", time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(feed.Items[0].Enclosures); got != maxEnclosures {
		t.Errorf("enclosures = %d, want %d", got, maxEnclosures)
	}
}

func TestParseStripsTrackingParamsFromLinks(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Blog</title>
    <item>
      <title>Post</title>
      <guid>post-1</guid>
      <description>&lt;a href="https://target.example/p?utm_source=x&amp;amp;id=1"&gt;go&lt;/a&gt;</description>
    </item>
  </channel>
</rss>`

	stripped, err := Parse([]byte(data), "https://example.com/feed.xml", testNow, true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stripped.Items[0].Content, "utm_source") ||
		!strings.Contains(stripped.Items[0].Content, "id=1") {
		t.Errorf("stripped content = %q", stripped.Items[0].Content)
	}

	kept, err := Parse([]byte(data), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(kept.Items[0].Content, "utm_source") {
		t.Errorf("strip=false content = %q", kept.Items[0].Content)
	}
}
