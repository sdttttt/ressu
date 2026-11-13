// Command pgo-collect exercises Ressu's core hot paths under a CPU profile
// to produce default.pgo for profile-guided optimization. CI runs it before
// packaging so every release builds with a fresh profile of its own code.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync/atomic"

	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/core/store"
)

var version atomic.Int64

func feedXML() string {
	version.Add(1)
	items := ""
	for i := range 40 {
		items += fmt.Sprintf(`<item><title>文章 %d r%d</title><link>https://example.com/%d</link>
<guid>guid-%d</guid><pubDate>Wed, 02 Oct 2024 12:%02d:00 GMT</pubDate>
<description>&lt;p&gt;第 %d 篇的摘要，足够长以模拟真实的摘要提取和清洗开销。&lt;/p&gt;</description>
<content:encoded>&lt;p&gt;这是第 %d 篇文章的完整正文。&lt;/p&gt;&lt;p&gt;包含多个段落，&lt;strong&gt;行内样式&lt;/strong&gt;，
以及一张 &lt;img src="https://example.com/pic.png"&gt; 图片，用来模拟真实订阅源的正文清洗与重写开销。&lt;/p&gt;</content:encoded>
</item>`, i, version.Load(), i, i, i%60, i, i)
	}
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/"><channel>
<title>测试源</title><link>https://example.com/</link><description>基准</description>
` + items + `</channel></rss>`
}

func main() {
	out := flag.String("out", "default.pgo", "where to write the CPU profile")
	flag.Parse()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	profile, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(profile); err != nil {
		log.Fatal(err)
	}
	defer func() {
		pprof.StopCPUProfile()
		_ = profile.Close()
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = io.WriteString(w, feedXML())
	}))
	defer server.Close()

	ctx := context.Background()
	for round := range 3 {
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("ressu-pgo-%d", round))
		_ = os.RemoveAll(dir)
		st, err := store.Open(dir)
		if err != nil {
			log.Fatal(err)
		}
		feeds := service.NewFeeds(st)
		posts := service.NewPosts(st)

		var feedIDs []int64
		for n := range 4 {
			feed, err := feeds.Add(ctx, "", fmt.Sprintf("%s/feed/%d", server.URL, n), "")
			if err != nil {
				log.Fatal(err)
			}
			feedIDs = append(feedIDs, feed.ID)
		}

		for refresh := range 12 {
			if _, err := feeds.RefreshAll(ctx); err != nil {
				log.Fatal(err)
			}
			for _, id := range feedIDs {
				page, err := posts.List(service.PostQuery{FeedID: id, Limit: 100})
				if err != nil {
					log.Fatal(err)
				}
				for _, summary := range page.Posts {
					if _, err := posts.Get(id, summary.GUID); err != nil {
						log.Fatal(err)
					}
					if refresh%3 == 0 {
						if _, err := posts.ToggleRead(id, summary.GUID); err != nil {
							log.Fatal(err)
						}
					}
				}
				if _, err := posts.MarkAllRead(id); err != nil {
					log.Fatal(err)
				}
			}
		}
		if _, err := feeds.BuildOPML(); err != nil {
			log.Fatal(err)
		}
		_ = st.Close()
		_ = os.RemoveAll(dir)
	}
}
