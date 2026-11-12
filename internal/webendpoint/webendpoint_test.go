package webendpoint

import (
	"context"
	"errors"
	"testing"
)

func TestStitchSnapshotRequiresCallerWindow(t *testing.T) {
	posts := NewPosts(nil)

	if err := posts.StitchSnapshot(context.Background(), "", nil); !errors.Is(err, ErrNoSnapshotWindow) {
		t.Fatalf("StitchSnapshot without caller = %v; want ErrNoSnapshotWindow", err)
	}
}
