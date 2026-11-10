package gui

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/store"
)

func TestNotificationsNormalizeAndCap(t *testing.T) {
	notifications := NewNotifications(nil)
	notifications.Show(Notification{Type: "WARNING", Title: "  刷新失败  "})

	if len(notifications.pending) != 1 {
		t.Fatalf("pending = %d; want 1", len(notifications.pending))
	}
	if got := notifications.pending[0]; got.Type != "warning" || got.Title != "刷新失败" {
		t.Fatalf("notification = %#v; want warning/刷新失败", got)
	}

	for i := range maxPendingNotifications + 5 {
		notifications.Show(Notification{Type: "info", Title: fmt.Sprint(i)})
	}
	if len(notifications.pending) != maxPendingNotifications {
		t.Fatalf("pending = %d; want %d", len(notifications.pending), maxPendingNotifications)
	}
	if got := notifications.pending[0].Title; got != "5" {
		t.Errorf("oldest retained notification = %q; want 5", got)
	}
}

func TestNotificationsDrawShowsToast(t *testing.T) {
	notifications := NewNotifications(func() (string, string) {
		return store.ThemeSepia, store.FontSizeLarge
	})
	notifications.Show(Notification{
		Type:        "success",
		Title:       "已导出订阅源",
		Description: "共 2 个订阅源",
	})

	tester := ui.NewTester(notifications.draw, 640, 480)
	for _, want := range []string{"已导出订阅源", "共 2 个订阅源", "×"} {
		if !tester.HasText(want) {
			t.Errorf("notification is missing %q", want)
		}
	}
}
