package service

import (
	"context"

	appevent "github.com/sdttttt/ressu/internal/event"
	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

// Status forwards page selection changes to native UI and other pages.
type Status struct {
	hub       *appevent.Hub
	pageReady func()
}

// NewStatus serves the page-driven application state.
func NewStatus(hub *appevent.Hub, pageReady func()) *Status {
	return &Status{hub: hub, pageReady: pageReady}
}

// SelectFeed records the feed the page selected; zero selects the aggregate
// list.
func (s *Status) SelectFeed(_ context.Context, feedID int64) {
	s.hub.SelectionChanged(feedID)
}

// SelectPost records the article a reader selected.
func (s *Status) SelectPost(_ context.Context, post store.PostSummary) {
	s.hub.PostSelected(
		appstatus.PostSelection{FeedID: post.FeedID, GUID: post.GUID},
		post,
	)
}

// PageReady lets the app show once the page has mounted its splash, then
// broadcasts the complete state to it.
func (s *Status) PageReady(_ context.Context) {
	if s.pageReady != nil {
		s.pageReady()
	}
	s.hub.PublishSnapshot()
}
