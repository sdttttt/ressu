// Package service exposes the app's data to the frontend through MyGo
// bindings, and tells pages when it changes through typed events.
package service

import (
	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/store"
)

// SyncState reports how a refresh is going.
type SyncState struct {
	Active bool   `json:"active"`
	Done   int    `json:"done"`
	Total  int    `json:"total"`
	FeedID int64  `json:"feedId"`
	Feed   string `json:"feed"`
	Error  string `json:"error,omitempty"`
}

// FeedsSnapshot is the subscription list with the total unread count.
type FeedsSnapshot struct {
	Feeds  []store.Feed `json:"feeds"`
	Unread int          `json:"unread"`
}

// PostsChanged says which feed's posts changed, and how many are new.
type PostsChanged struct {
	FeedID int64 `json:"feedId"`
	New    int   `json:"new"`
}

// Events reach the pages of every window.
var (
	// SyncStateEvent reports the progress of a refresh.
	SyncStateEvent = mygo.NewEvent[SyncState]("sync:state")
	// FeedsChanged carries the subscription list once it changes.
	FeedsChanged = mygo.NewEvent[FeedsSnapshot]("feeds:changed")
	// PostsChangedEvent says which feed's posts changed.
	PostsChangedEvent = mygo.NewEvent[PostsChanged]("posts:changed")
	// Command carries a menu command for the page to run.
	Command = mygo.NewEvent[string]("ui:command")
)

// Commands the menu sends to the page.
const (
	CommandAddFeed  = "add-feed"
	CommandRefresh  = "refresh"
	CommandSettings = "settings"
)

// broadcastFeeds tells the pages about the current subscriptions.
func broadcastFeeds(st *store.Store) {
	_ = FeedsChanged.Broadcast(FeedsSnapshot{Feeds: st.Feeds(), Unread: st.Unread()})
}
