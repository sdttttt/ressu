package service

import (
	"context"
	"fmt"
	"os"

	"github.com/egoist/mygo"

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

// ImportOPML shows a native file picker, reads the chosen OPML file and
// subscribes to every feed it lists. Duplicates and unreachable feeds are
// counted but do not stop the import.
func (f *Feeds) ImportOPML(ctx context.Context) (ImportSummary, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title: "选择 OPML 文件",
		Filters: []mygo.FileFilter{
			{Name: "OPML 订阅列表", Extensions: []string{"opml", "xml"}},
		},
	})
	if err != nil {
		return ImportSummary{}, err
	}
	if len(paths) == 0 {
		return ImportSummary{}, nil
	}

	file, err := os.Open(paths[0])
	if err != nil {
		return ImportSummary{}, err
	}
	defer file.Close()

	urls, err := rss.ReadOPML(file)
	if err != nil {
		return ImportSummary{}, err
	}

	return f.importOPMLURLs(ctx, urls)
}

// ExportOPML shows a native save dialog and writes every subscription to the
// chosen OPML file, whose path it returns. A cancelled dialog returns no path.
func (f *Feeds) ExportOPML(ctx context.Context) (string, error) {
	data, err := rss.ExportOPML(opmlFeeds(f.store.Feeds()))
	if err != nil {
		return "", err
	}

	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      mygo.CallerWindow(ctx),
		Title:       "导出 OPML 文件",
		DefaultPath: "ressu-subscriptions.opml",
		Filters: []mygo.FileFilter{
			{Name: "OPML 订阅列表", Extensions: []string{"opml"}},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("could not save the OPML file: %w", err)
	}
	return path, nil
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

// importOPMLURLs subscribes to every URL in the list, counting results.
func (f *Feeds) importOPMLURLs(ctx context.Context, urls []string) (ImportSummary, error) {
	var summary ImportSummary
	for _, feedURL := range urls {
		if _, err := f.Add(ctx, "", feedURL); err != nil {
			if err == store.ErrDuplicate {
				summary.Skipped++
			} else {
				summary.Failed++
				summary.Errors = append(summary.Errors, feedURL+": "+err.Error())
			}
			continue
		}
		summary.Imported++
	}
	return summary, nil
}
