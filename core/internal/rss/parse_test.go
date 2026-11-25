package rss

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestParseRSS2(t *testing.T) {
	feed, err := Parse([]byte(readFixture(t, "rss2.xml")), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}

	if feed.Meta.Title != "Example Feed" {
		t.Errorf("title = %q", feed.Meta.Title)
	}
	if feed.Meta.SiteURL != "https://example.com/" {
		t.Errorf("site url = %q", feed.Meta.SiteURL)
	}
	if feed.Meta.IconURL != "https://example.com/icon.png" {
		t.Errorf("icon = %q", feed.Meta.IconURL)
	}
	if feed.Meta.Language != "en-us" {
		t.Errorf("language = %q", feed.Meta.Language)
	}
	if feed.Meta.Description != "A sample feed" {
		t.Errorf("description = %q", feed.Meta.Description)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("items = %d", len(feed.Items))
	}

	first := feed.Items[0]
	if first.GUID != "post-1" {
		t.Errorf("guid = %q", first.GUID)
	}
	if first.URL != "https://example.com/posts/1" {
		t.Errorf("url = %q", first.URL)
	}
	if first.Summary != "Hello world." {
		t.Errorf("summary = %q", first.Summary)
	}
	if !strings.Contains(first.Content, "<strong>world</strong>") {
		t.Errorf("content = %q", first.Content)
	}
	if strings.Contains(first.Content, "<script") || strings.Contains(first.Content, "alert(1)") {
		t.Errorf("content = %q", first.Content)
	}
	if first.WordCount != 2 {
		t.Errorf("word count = %d, want 2", first.WordCount)
	}
	want := time.Date(2024, 10, 2, 12, 0, 0, 0, time.UTC)
	if !first.PublishedAt.Equal(want) {
		t.Errorf("published = %s, want %s", first.PublishedAt, want)
	}
	if first.Author == "" {
		t.Error("author is empty")
	}

	second := feed.Items[1]
	if second.URL != "https://example.com/posts/2" {
		t.Errorf("relative link resolved to %q", second.URL)
	}
	if second.GUID != second.URL {
		t.Errorf("guid without one should fall back to the link, got %q", second.GUID)
	}
	if second.Summary != "Plain text & more" {
		t.Errorf("summary = %q", second.Summary)
	}
	if second.WordCount != 3 {
		t.Errorf("word count = %d, want 3", second.WordCount)
	}
}

func TestParseAtom(t *testing.T) {
	feed, err := Parse([]byte(readFixture(t, "atom.xml")), "https://atom.example/feed", testNow, false)
	if err != nil {
		t.Fatal(err)
	}

	if feed.Meta.Title != "Atom Example" {
		t.Errorf("title = %q", feed.Meta.Title)
	}
	if feed.Meta.SiteURL != "https://atom.example/" {
		t.Errorf("site url = %q", feed.Meta.SiteURL)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("items = %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.GUID != "urn:uuid:entry-1" {
		t.Errorf("guid = %q", item.GUID)
	}
	if item.Author != "Ada" {
		t.Errorf("author = %q", item.Author)
	}
	want := time.Date(2024, 10, 4, 8, 0, 0, 0, time.UTC)
	if !item.PublishedAt.Equal(want) {
		t.Errorf("published = %s", item.PublishedAt)
	}
	if !strings.Contains(item.Content, "<p>Body &amp; more</p>") {
		t.Errorf("content = %q", item.Content)
	}
	if item.Summary != "Short summary" {
		t.Errorf("summary = %q", item.Summary)
	}
}

func TestParseJSONFeed(t *testing.T) {
	feed, err := Parse([]byte(readFixture(t, "jsonfeed.json")), "https://json.example/feed.json", testNow, false)
	if err != nil {
		t.Fatal(err)
	}

	if feed.Meta.Title != "JSON Example" {
		t.Errorf("title = %q", feed.Meta.Title)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("items = %d", len(feed.Items))
	}
	item := feed.Items[0]
	if item.GUID != "j1" || item.Author != "Jo" {
		t.Errorf("item = %+v", item)
	}
	if item.Summary != "Hi there" {
		t.Errorf("summary = %q", item.Summary)
	}
}

func TestParseBrokenFeed(t *testing.T) {
	if _, err := Parse([]byte(readFixture(t, "broken.xml")), "https://example.com/", testNow, false); err == nil {
		t.Fatal("a page that is not a feed should fail")
	}
}

func TestParseFallsBackToTitleAndContent(t *testing.T) {
	text := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>No ids</title>
<item><title>Nameless</title></item>
</channel></rss>`

	feed, err := Parse([]byte(text), "https://example.com/feed", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	item := feed.Items[0]
	if !strings.HasPrefix(item.GUID, "Nameless|") || strings.Count(item.GUID, "|") != 1 {
		t.Errorf("guid = %q, want Nameless|<8 hex digits>", item.GUID)
	}
	if !item.PublishedAt.Equal(testNow) {
		t.Errorf("published = %s", item.PublishedAt)
	}
}

func TestPlainText(t *testing.T) {
	cases := []struct {
		name   string
		markup string
		want   string
	}{
		{"tags", "<p>Hello <em>there</em></p>", "Hello there"},
		{"entities", "Tom &amp; Jerry &lt;3", "Tom & Jerry <3"},
		{"script", `keep<script>alert("x")</script>drop`, "keepdrop"},
		{"style", "a<style>p{color:red}</style>b", "ab"},
		{"whitespace", "  a\n\t b  ", "a b"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := plainText(c.markup)
			if got != c.want {
				t.Errorf("plainText(%q) = %q, want %q", c.markup, got, c.want)
			}
		})
	}
}

func TestPlainTextTruncates(t *testing.T) {
	got := plainText(strings.Repeat("字", SummaryLimit+10))
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("summary should be cut, got %q", got)
	}
	if n := len([]rune(got)); n != SummaryLimit+1 {
		t.Errorf("rune count = %d", n)
	}
}

func TestParseDropsNonHTTPLink(t *testing.T) {
	raw := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>F</title>
<item><title>a</title><link>javascript:alert(1)</link></item>
<item><title>b</title><link>https://example.com/posts/1</link></item>
</channel></rss>`
	feed, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Items[0].URL != "" {
		t.Errorf("javascript link survived as %q", feed.Items[0].URL)
	}
	if feed.Items[1].URL != "https://example.com/posts/1" {
		t.Errorf("https link = %q", feed.Items[1].URL)
	}
}

func TestParseFallbackGUIDSeparatesSameBatchItems(t *testing.T) {
	raw := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>F</title>
<item><title>same</title><description>first</description></item>
<item><title>same</title><description>second</description></item>
</channel></rss>`
	feed, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Items[0].GUID == feed.Items[1].GUID {
		t.Errorf("same-batch items without link or date share a GUID: %q", feed.Items[0].GUID)
	}
}

func TestParseFallbackGUIDIsStableAcrossRefreshes(t *testing.T) {
	raw := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>F</title>
<item><title>undated</title><description>body</description></item>
</channel></rss>`
	first, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	next, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow.Add(time.Hour), false)
	if err != nil {
		t.Fatal(err)
	}

	// Only the fetch time dates the item. Carrying it would store a new copy
	// on every refresh.
	guid := first.Items[0].GUID
	if guid != next.Items[0].GUID {
		t.Errorf("fallback GUID changes with the fetch time: %q, then %q", guid, next.Items[0].GUID)
	}
	if title, hash, found := strings.Cut(guid, "|"); !found || title != "undated" ||
		len(hash) != 8 || strings.Contains(hash, "|") {
		t.Errorf("fallback GUID = %q, want undated|<8 hex digits>", guid)
	}
	if !first.Items[0].PublishedAt.Equal(testNow) || !next.Items[0].PublishedAt.Equal(testNow.Add(time.Hour)) {
		t.Error("an undated item should be dated by its fetch")
	}
}

func TestParseFallbackGUIDKeepsTheDate(t *testing.T) {
	raw := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>F</title>
<item><title>dated</title><description>body</description>
<pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate></item>
</channel></rss>`
	feed, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}

	if guid := feed.Items[0].GUID; !strings.HasPrefix(guid, "dated|2024-01-01T00:00:00Z|") {
		t.Errorf("fallback GUID = %q, want the date after the title", guid)
	}
}

func TestValidateURL(t *testing.T) {
	valid := []string{"https://example.com/feed", "http://example.com/rss.xml"}
	for _, raw := range valid {
		if err := ValidateURL(raw); err != nil {
			t.Errorf("%q should be valid: %v", raw, err)
		}
	}
	invalid := []string{"", "example.com/feed", "ftp://example.com/feed", "file:///etc/passwd"}
	for _, raw := range invalid {
		if err := ValidateURL(raw); err == nil {
			t.Errorf("%q should be invalid", raw)
		}
	}
}

func TestParseExtractsFirstImage(t *testing.T) {
	raw := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>F</title>
<item><title>with image</title><link>https://example.com/posts/1</link>
<description>&lt;p&gt;text&lt;/p&gt;&lt;img src="/pic/cover.png"&gt;</description></item>
<item><title>text only</title><link>https://example.com/posts/2</link>
<description>&lt;p&gt;plain&lt;/p&gt;</description></item>
</channel></rss>`
	feed, err := Parse([]byte(raw), "https://example.com/feed.xml", testNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := feed.Items[0].Image; got != "https://example.com/pic/cover.png" {
		t.Errorf("image = %q", got)
	}
	if got := feed.Items[1].Image; got != "" {
		t.Errorf("image = %q, want empty", got)
	}
}
