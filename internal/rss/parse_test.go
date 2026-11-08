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
	feed, err := Parse([]byte(readFixture(t, "rss2.xml")), "https://example.com/feed.xml", testNow)
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
}

func TestParseAtom(t *testing.T) {
	feed, err := Parse([]byte(readFixture(t, "atom.xml")), "https://atom.example/feed", testNow)
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
	feed, err := Parse([]byte(readFixture(t, "jsonfeed.json")), "https://json.example/feed.json", testNow)
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
	if _, err := Parse([]byte(readFixture(t, "broken.xml")), "https://example.com/", testNow); err == nil {
		t.Fatal("a page that is not a feed should fail")
	}
}

func TestParseFallsBackToTitleAndDate(t *testing.T) {
	text := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>No ids</title>
<item><title>Nameless</title></item>
</channel></rss>`

	feed, err := Parse([]byte(text), "https://example.com/feed", testNow)
	if err != nil {
		t.Fatal(err)
	}
	item := feed.Items[0]
	if item.GUID != "Nameless|"+testNow.Format(time.RFC3339Nano) {
		t.Errorf("guid = %q", item.GUID)
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
			got := plainText(c.markup, SummaryLimit)
			if got != c.want {
				t.Errorf("plainText(%q) = %q, want %q", c.markup, got, c.want)
			}
		})
	}
}

func TestPlainTextTruncates(t *testing.T) {
	got := plainText(strings.Repeat("字", SummaryLimit+10), SummaryLimit)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("summary should be cut, got %q", got)
	}
	if n := len([]rune(got)); n != SummaryLimit+1 {
		t.Errorf("rune count = %d", n)
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
