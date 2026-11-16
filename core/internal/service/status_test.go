package service

import (
	"context"
	"testing"

	appevent "github.com/sdttttt/ressu/core/internal/event"
	appstatus "github.com/sdttttt/ressu/core/internal/status"
	"github.com/sdttttt/ressu/core/internal/store"
)

func TestPageReadyRestoresSelectedPost(t *testing.T) {
	var events []appevent.Event
	hub := appevent.NewHub(appstatus.New("test"), func(e appevent.Event) error {
		events = append(events, e)
		return nil
	})
	status := NewStatus(hub, nil)

	status.PageReady(context.Background())
	if len(events) != 1 || events[0].EID != appevent.EventStateSnapshot {
		t.Fatalf("events without selection = %+v", events)
	}

	post := store.PostSummary{FeedID: 7, GUID: "guid-1", Title: "文章"}
	status.SelectPost(context.Background(), post, 1)
	events = events[:0]

	status.PageReady(context.Background())
	if len(events) != 2 || events[0].EID != appevent.EventStateSnapshot ||
		events[1].EID != appevent.EventPostSelected {
		t.Fatalf("events with selection = %+v", events)
	}
	if events[1].Post == nil || events[1].Post.GUID != "guid-1" {
		t.Errorf("restored post = %+v", events[1].Post)
	}
	if events[1].Direction != 0 {
		t.Errorf("restored direction = %d", events[1].Direction)
	}
}
