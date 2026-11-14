package viewer

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func TestSnapshotViewerRunsSaveAndCopy(t *testing.T) {
	data, err := testViewerPNG()
	if err != nil {
		t.Fatal(err)
	}
	content, err := newSnapshotViewerContent(data, " 测试  文章 ", nil, SnapshotCallbacks{
		Save: func(parent *mygo.Window, gotData []byte, title string) (string, error) {
			if parent != nil || !slices.Equal(gotData, data) || title != "测试 文章" {
				t.Fatalf("save callback = parent %v, %d bytes, title %q", parent, len(gotData), title)
			}
			return "test.png", nil
		},
		Copy: func(gotData []byte) error {
			if !slices.Equal(gotData, data) {
				t.Fatal("copy callback received different image bytes")
			}
			return nil
		},
		Notify: func(level, message string) {
			t.Logf("%s: %s", level, message)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	tester := ui.NewTester(func(c *ui.Context) { content.draw(c) }, 960, 640)
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if err := tester.Click("复制"); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotViewerReportsFailures(t *testing.T) {
	data, err := testViewerPNG()
	if err != nil {
		t.Fatal(err)
	}
	var notifications []string
	content, err := newSnapshotViewerContent(data, "", nil, SnapshotCallbacks{
		Save: func(*mygo.Window, []byte, string) (string, error) {
			return "", errors.New("save failed")
		},
		Copy: func([]byte) error {
			return errors.New("copy failed")
		},
		Notify: func(level, message string) {
			notifications = append(notifications, level+":"+message)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	tester := ui.NewTester(func(c *ui.Context) { content.draw(c) }, 960, 640)
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if err := tester.Click("复制"); err != nil {
		t.Fatal(err)
	}
	if got := notifications; len(got) != 2 || got[0] != "error:无法保存快照" ||
		got[1] != "error:无法复制快照" {
		t.Fatalf("notifications = %q", got)
	}
}

func TestSnapshotTitleTrimsAndLimitsText(t *testing.T) {
	if got := snapshotTitle("  spaced   title  "); got != "spaced title" {
		t.Fatalf("snapshotTitle() = %q", got)
	}
	if got := snapshotTitle(""); got != "文章快照" {
		t.Fatalf("empty snapshotTitle() = %q", got)
	}

	long := strings.Repeat("文", snapshotTitleRunes+10)
	got := []rune(snapshotTitle(long))
	if len(got) != snapshotTitleRunes+1 || string(got[:snapshotTitleRunes]) != strings.Repeat("文", snapshotTitleRunes) ||
		string(got[snapshotTitleRunes]) != "…" {
		t.Fatalf("long snapshot title has %d runes", len(got))
	}
}
