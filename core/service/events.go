// Package service keeps the reader's business logic: subscriptions, posts,
// preferences and snapshots. It is UI-framework agnostic; state changes are
// published as domain events through the event Hub.
package service

import (
	appevent "github.com/sdttttt/ressu/core/event"
	appstatus "github.com/sdttttt/ressu/core/status"
)

// SyncState reports how a refresh is going.
type SyncState = appstatus.SyncState

// FeedsSnapshot is the subscription list with the total unread count.
type FeedsSnapshot = appstatus.FeedsSnapshot

// PostsChanged says which feed's posts changed, and how many are new.
type PostsChanged = appevent.PostsChanged
