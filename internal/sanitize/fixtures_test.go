package sanitize_test

import (
	"embed"
	"strings"
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/rss"
	"github.com/sdttttt/ressu/internal/sanitize"
)

// The fixtures are complete responses from feeds that had previously failed
// or produced overly permissive article HTML in production.
//
//go:embed 12306.xml 3dm.xml els.xml yystv.xml
var realFeedFixtures embed.FS

func TestHTMLCleansRealRSSFeeds(t *testing.T) {
	fixtures := []struct {
		name            string
		feedURL         string
		firstTitle      string
		firstContentTag string
	}{
		{
			name:            "12306.xml",
			feedURL:         "https://www.12306.cn/mormhweb/zxdt/index_zxdt.html",
			firstTitle:      "公 告",
			firstContentTag: `<h3 class="title">公`,
		},
		{
			name:            "3dm.xml",
			feedURL:         "https://www.3dmgame.com/news/",
			firstTitle:      "蕾卡的移动工坊",
			firstContentTag: `src="https://img.3dmgame.com/`,
		},
		{
			name:            "els.xml",
			feedURL:         "https://sputniknews.com/news",
			firstTitle:      "Oil Up 2nd Week in Row",
			firstContentTag: `<div class="article__header">`,
		},
		{
			name:            "yystv.xml",
			feedURL:         "https://www.yystv.cn",
			firstTitle:      "光芒之池",
			firstContentTag: `src="https://alioss.yystv.cn/`,
		},
	}
	now := time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			data, err := realFeedFixtures.ReadFile(fixture.name)
			if err != nil {
				t.Fatal(err)
			}

			feed, err := rss.Parse(data, fixture.feedURL, now)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if feed == nil || len(feed.Items) == 0 {
				t.Fatalf("Parse() returned %d items, want at least one", len(feed.Items))
			}

			if !strings.Contains(feed.Items[0].Title, fixture.firstTitle) ||
				!strings.Contains(feed.Items[0].Content, fixture.firstContentTag) {
				t.Fatalf("fixture article markup was not preserved: %s", truncate(feed.Items[0].Content, 300))
			}

			for i, item := range feed.Items {
				cleaned := sanitize.HTML(item.Content, item.URL)
				if cleaned != item.Content {
					t.Fatalf("item %d is not idempotent:\nstored = %q\ncleaned = %q", i, item.Content, cleaned)
				}

				lower := strings.ToLower(cleaned)
				for _, unsafe := range []string{
					"<script", "<style", "<iframe", "<object", "<embed", "<form",
					"<input", "<button", "javascript:", "onclick=", "onerror=", "onload=",
				} {
					if strings.Contains(lower, unsafe) {
						t.Fatalf("item %d contains %q: %s", i, unsafe, truncate(cleaned, 300))
					}
				}
			}
		})
	}
}

func TestHTMLCleansReal12306Article(t *testing.T) {
	data, err := realFeedFixtures.ReadFile("12306.xml")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)
	feed, err := rss.Parse(data, "https://www.12306.cn/mormhweb/zxdt/index_zxdt.html", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Items) != 20 {
		t.Fatalf("items = %d, want 20", len(feed.Items))
	}

	if !feed.Items[1].PublishedAt.Equal(now) || !feed.Items[2].PublishedAt.Equal(now) {
		t.Fatalf("invalid dates = %s, %s; want %s", feed.Items[1].PublishedAt, feed.Items[2].PublishedAt, now)
	}

	var article *rss.Item
	for i := range feed.Items {
		if strings.Contains(feed.Items[i].Title, "起售时间") {
			article = &feed.Items[i]
			break
		}
	}
	if article == nil {
		t.Fatal("ticket sale time article not found")
	}

	lower := strings.ToLower(article.Content)
	for _, removed := range []string{
		"<style", ".trs_preappend", "padding-top:", "color:black", "<link",
		"<input", "<button", "<!--",
	} {
		if strings.Contains(lower, removed) {
			t.Fatalf("%q survived: %s", removed, truncate(article.Content, 500))
		}
	}
	for _, wanted := range []string{
		"自2014年11月28日起",
		"8：00 起售车站",
		"北京西、南京",
	} {
		if !strings.Contains(article.Content, wanted) {
			t.Fatalf("article text missing %q", wanted)
		}
	}
}

func TestHTMLCleansRealSputnikArticle(t *testing.T) {
	data, err := realFeedFixtures.ReadFile("els.xml")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)
	feed, err := rss.Parse(data, "https://sputniknews.com/news", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Items) != 14 {
		t.Fatalf("items = %d, want 14", len(feed.Items))
	}

	article := feed.Items[0]
	if !strings.Contains(article.Title, "Oil Up 2nd Week in Row") {
		t.Fatalf("title = %q", article.Title)
	}

	lower := strings.ToLower(article.Content)
	for _, removed := range []string{
		"srcset=", "style=", "width=", "height=", "referrerpolicy=", "<noindex",
	} {
		if strings.Contains(lower, removed) {
			t.Fatalf("%q survived: %s", removed, truncate(article.Content, 500))
		}
	}
	for _, wanted := range []string{
		`src="https://cdnn1.img.sputniknews.com/`,
		`loading="lazy"`,
		`decoding="async"`,
		`<div class="article__body">`,
		"Brent crude, the London-traded global benchmark for oil",
	} {
		if !strings.Contains(article.Content, wanted) {
			t.Fatalf("article missing %q: %s", wanted, truncate(article.Content, 500))
		}
	}
}

func TestHTMLCleansRealYYSTVArticle(t *testing.T) {
	data, err := realFeedFixtures.ReadFile("yystv.xml")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)
	feed, err := rss.Parse(data, "https://www.yystv.cn", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Items) == 0 {
		t.Fatalf("Parse() returned %d items, want at least one", len(feed.Items))
	}

	article := feed.Items[0]
	if !strings.Contains(article.Title, "光芒之池") {
		t.Fatalf("title = %q", article.Title)
	}
	for _, wanted := range []string{
		`src="https://alioss.yystv.cn/`,
		`<a href="https://www.youtube.com/`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
	} {
		if !strings.Contains(article.Content, wanted) {
			t.Fatalf("content missing %q: %s", wanted, truncate(article.Content, 300))
		}
	}
	if strings.Contains(strings.ToLower(article.Content), "style=") {
		t.Fatalf("inline style survived: %s", truncate(article.Content, 300))
	}
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "…"
}
