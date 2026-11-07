package service

import (
	"context"
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

// importOPMLURLs subscribes to every URL in the list, counting results.
func (f *Feeds) importOPMLURLs(ctx context.Context, urls []string) (ImportSummary, error) {
	var summary ImportSummary
	for _, feedURL := range urls {
		if _, err := f.Add(ctx, feedURL); err != nil {
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
