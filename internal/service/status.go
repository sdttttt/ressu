package service

import (
	"context"

	appstatus "github.com/sdttttt/ressu/internal/status"
)

// Status forwards page selection changes to the native status bar.
type Status struct {
	model     *appstatus.Model
	pageReady func()
}

// NewStatus serves the native status bar's page-driven state.
func NewStatus(model *appstatus.Model, pageReady func()) *Status {
	return &Status{model: model, pageReady: pageReady}
}

// SelectFeed tells the native status bar which feed the page selected; zero
// selects the aggregate list.
func (s *Status) SelectFeed(_ context.Context, feedID int64) {
	s.model.SetCurrentFeed(feedID)
}

// PageReady lets the app show once the page has mounted its splash.
func (s *Status) PageReady(_ context.Context) {
	if s.pageReady != nil {
		s.pageReady()
	}
}
