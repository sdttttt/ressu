package action

import (
	"bytes"
	"image"
	"image/png"
	"path/filepath"
	"testing"

	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/core/store"
)

func TestStitchSnapshotOpensNativeViewer(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})

	ui := &fakeUI{}
	ui.savePath = filepath.Join(t.TempDir(), "snapshot.png")
	actions := New(nil, service.NewPosts(st), service.NewSettings(st), nil, nil, ui)
	capture := testSnapshotCapture(t, 4, 2)
	tiles := []service.SnapshotTile{{
		Offset: 0,
		PNG:    capture,
		Geometry: service.SnapshotGeometry{
			ViewportWidth:  4,
			ViewportHeight: 2,
			CardWidth:      2,
			ContentHeight:  2,
		},
	}}

	if err := actions.stitchSnapshot(WindowHandle{}, "文章 标题", tiles); err != nil {
		t.Fatal(err)
	}
	if len(ui.snapshotData) == 0 || ui.snapshotTitle != "文章 标题" {
		t.Fatalf("snapshot viewer = %d bytes, title %q", len(ui.snapshotData), ui.snapshotTitle)
	}

	path, err := actions.SaveSnapshot(ui.snapshotParent, ui.snapshotTitle, ui.snapshotData)
	if err != nil || path != ui.savePath {
		t.Fatalf("viewer save = %q, %v", path, err)
	}
}

func TestStitchSnapshotRejectsInvalidTiles(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})

	ui := &fakeUI{}
	actions := New(nil, service.NewPosts(st), service.NewSettings(st), nil, nil, ui)
	if err := actions.stitchSnapshot(WindowHandle{}, "", nil); err == nil {
		t.Fatal("stitchSnapshot(WindowHandle{}, \"\", nil) did not fail")
	}
	if ui.snapshotData != nil {
		t.Error("invalid captures opened the snapshot viewer")
	}
}

func testSnapshotCapture(t *testing.T, width, height int) []byte {
	t.Helper()
	var output bytes.Buffer
	if err := png.Encode(&output, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
