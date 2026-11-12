package service

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sdttttt/ressu/core/rss"
)

func TestImportOPMLURLs(t *testing.T) {
	feedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(feedXML(itemXML("a", "第一篇"))))
	})
	ok1 := httptest.NewServer(feedHandler)
	defer ok1.Close()
	ok2 := httptest.NewServer(feedHandler)
	defer ok2.Close()

	st := openStore(t)
	feeds := NewFeeds(st)

	if _, err := feeds.Add(t.Context(), "", ok1.URL); err != nil {
		t.Fatalf("pre-add: %v", err)
	}

	opml := "<opml><body>" +
		"<outline type=\"rss\" xmlUrl=\"" + ok1.URL + "\"/>" +
		"<outline type=\"rss\" xmlUrl=\"" + ok2.URL + "\"/>" +
		"<outline type=\"rss\" xmlUrl=\"https://gone.example/feed.xml\"/>" +
		"</body></opml>"
	urls, err := rss.ParseOPML([]byte(opml))
	if err != nil {
		t.Fatalf("ParseOPML: %v", err)
	}

	summary, err := feeds.importOPMLURLs(t.Context(), urls)
	if err != nil {
		t.Fatalf("importOPMLURLs: %v", err)
	}
	if summary.Imported != 1 {
		t.Errorf("Imported = %d, want 1", summary.Imported)
	}
	if summary.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", summary.Skipped)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", summary.Failed)
	}
}

func TestImportOPMLEmptyURLs(t *testing.T) {
	st := openStore(t)
	feeds := NewFeeds(st)
	summary, err := feeds.importOPMLURLs(t.Context(), nil)
	if err != nil {
		t.Fatalf("importOPMLURLs: %v", err)
	}
	if summary.Imported != 0 {
		t.Errorf("Imported = %d, want 0", summary.Imported)
	}
	if len(st.Feeds()) != 0 {
		t.Errorf("feeds = %d, want 0", len(st.Feeds()))
	}
}

func TestReadOPMLLimits(t *testing.T) {
	big := "<opml><body>" + strings.Repeat("<outline text=\"x\"/>", 1<<22) + "</body></opml>"
	if _, err := rss.ReadOPML(strings.NewReader(big)); err == nil {
		t.Error("ReadOPML(oversized) should fail")
	}
}

func TestImportOPMLStoresFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(feedXML(itemXML("a", "第一篇"))))
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)
	summary, err := feeds.importOPMLURLs(t.Context(), []string{server.URL})
	if err != nil {
		t.Fatalf("importOPMLURLs: %v", err)
	}
	if summary.Imported != 1 {
		t.Fatalf("Imported = %d, want 1", summary.Imported)
	}
	list := st.Feeds()
	if len(list) != 1 {
		t.Fatalf("feeds = %d, want 1", len(list))
	}
	if list[0].URL != server.URL {
		t.Errorf("URL = %q, want %q", list[0].URL, server.URL)
	}
}

func TestImportOPMLContextCancelled(t *testing.T) {
	st := openStore(t)
	feeds := NewFeeds(st)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	summary, _ := feeds.importOPMLURLs(ctx, []string{"https://a.example/feed.xml"})
	if summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", summary.Failed)
	}
}

func TestExportOPMLData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(feedXML(itemXML("a", "第一篇"))))
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)
	if _, err := feeds.Add(t.Context(), "", server.URL); err != nil {
		t.Fatalf("Add: %v", err)
	}

	data, err := rss.ExportOPML(opmlFeeds(st.Feeds()))
	if err != nil {
		t.Fatalf("ExportOPML: %v", err)
	}

	var doc struct {
		Body struct {
			Outlines []struct {
				Text        string `xml:"text,attr"`
				XMLURL      string `xml:"xmlUrl,attr"`
				HTMLURL     string `xml:"htmlUrl,attr"`
				Description string `xml:"description,attr"`
				Language    string `xml:"language,attr"`
			} `xml:"outline"`
		} `xml:"body"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal export: %v", err)
	}
	outlines := doc.Body.Outlines
	if len(outlines) != 1 {
		t.Fatalf("outlines = %d, want 1", len(outlines))
	}
	outline := outlines[0]
	if outline.Text != "测试源" ||
		outline.XMLURL != server.URL ||
		outline.HTMLURL != "https://example.com/" ||
		outline.Description != "一个测试用的订阅源" ||
		outline.Language != "" {
		t.Errorf("outline = %+v, want metadata from %s", outline, server.URL)
	}
}
