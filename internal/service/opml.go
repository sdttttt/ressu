package service

import (
	"github.com/sdttttt/ressu/internal/rss"
	"github.com/sdttttt/ressu/internal/store"
)

// ImportSummary reports what an OPML import did.
type ImportSummary struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors"`
}

// opmlFeeds converts stored subscriptions to OPML outlines.
func opmlFeeds(feeds []store.Feed) []rss.OPMLFeed {
	outlines := make([]rss.OPMLFeed, 0, len(feeds))
	for _, feed := range feeds {
		outlines = append(outlines, rss.OPMLFeed{
			Title:       feed.Title,
			URL:         feed.URL,
			SiteURL:     feed.SiteURL,
			Description: feed.Description,
			Language:    feed.Language,
		})
	}
	return outlines
}
