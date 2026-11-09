// Package service exposes the app's data to the frontend through MyGo
// bindings, and tells pages when it changes through events.
package service

import (
	appevent "github.com/sdttttt/ressu/internal/event"
	appstatus "github.com/sdttttt/ressu/internal/status"
)

// SyncState reports how a refresh is going.
type SyncState = appstatus.SyncState

// FeedsSnapshot is the subscription list with the total unread count.
type FeedsSnapshot = appstatus.FeedsSnapshot

// PostsChanged says which feed's posts changed, and how many are new.
type PostsChanged = appevent.PostsChanged

// Commands the menu sends to the page.
const (
	CommandAddFeed  = "add-feed"
	CommandRefresh  = "refresh"
	CommandSettings = "settings"
)
