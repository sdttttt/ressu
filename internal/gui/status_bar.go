// Package gui draws the parts of the reader that MyGO renders directly.
package gui

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/core/status"
	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
)

// StatusBar draws the reader's bottom status bar with MyGO native UI.
type StatusBar struct {
	model   *appstatus.Model
	onAbout func()
}

// NewStatusBar draws the state kept by model.
func NewStatusBar(model *appstatus.Model, onAbout func()) *StatusBar {
	return &StatusBar{model: model, onAbout: onAbout}
}

// View builds one frame of the status bar.
func (b *StatusBar) View(c *ui.Context) {
	snapshot := b.model.Snapshot()
	theme := guitheme.New(c, snapshot.Theme, snapshot.FontSize)
	fontScale := guitheme.FontScale(snapshot.FontSize)
	height := guitheme.StatusBarContentHeight * fontScale
	spacing := 12 * fontScale
	gap := 8 * fontScale
	indicator := 12 * fontScale

	ui.Column(c).Shrink(0).Children(func() {
		ui.Box(c).Height(1).Shrink(0).Background(theme.Border)
		ui.Themed(c, theme).Height(height).Shrink(0).Children(func() {
			ui.Row(c).Height(height).Padding(0, spacing).Gap(gap).AlignItems(ui.Center).
				Role(ui.RoleStatus).Label("应用状态").Children(func() {
				if snapshot.Sync.Active {
					ui.Spinner(c).Size(indicator, indicator).Label("正在刷新")
					ui.Textf(c, "正在刷新 %d/%d", snapshot.Sync.Done, snapshot.Sync.Total).
						TextColor(theme.TextMuted).FontFeatures("tnum").
						MinWidth(0).SingleLine()
					if snapshot.Sync.Feed != "" {
						ui.Textf(c, "· %s", snapshot.Sync.Feed).
							TextColor(theme.TextMuted).
							MinWidth(0).SingleLine()
					}
				} else {
					ui.Textf(c, "%d 订阅源", snapshot.FeedCount).
						TextColor(theme.TextMuted).FontFeatures("tnum")
					divider(c, theme, fontScale)
					ui.Badge(c, fmt.Sprintf("%d 未读", snapshot.Unread)).
						Background(theme.Surface).TextColor(theme.Text)
				}

				ui.Row(c).Grow(1)

				if snapshot.Sync.Active && snapshot.Sync.Error != "" {
					ui.Text(c, "刷新失败").TextColor(theme.Danger).
						Tooltip(snapshot.Sync.Error).SingleLine()
				}
				if snapshot.CurrentFeed != nil && snapshot.CurrentFeed.LastFetchedAt != nil {
					if snapshot.Sync.Active || snapshot.Sync.Error != "" {
						divider(c, theme, fontScale)
					}
					fetched := *snapshot.CurrentFeed.LastFetchedAt
					ui.Textf(c, "上次刷新 %s", relativeTime(snapshot.RelativeStart, fetched)).
						TextColor(theme.TextMuted).FontFeatures("tnum").SingleLine()
					scheduleRelativeRefresh(c, snapshot.RelativeStart)
				}
				if snapshot.Version != "" {
					if snapshot.CurrentFeed != nil && snapshot.CurrentFeed.LastFetchedAt != nil {
						divider(c, theme, fontScale)
					}
					version := ui.Textf(c, "v%s", snapshot.Version).
						TextColor(theme.TextMuted).FontFeatures("tnum")
					if b.onAbout != nil {
						version.Cursor(ui.CursorPointer).Tooltip("关于 Ressu").
							Focusable().OnClick(b.onAbout)
					}
				}
			})
		})
	})
}

func divider(c *ui.Context, theme *ui.Theme, fontScale float32) {
	ui.Box(c).Width(1).Height(14 * fontScale).Shrink(0).Background(theme.Border).
		Role(ui.RoleSeparator).Label("分隔线")
}

func scheduleRelativeRefresh(c *ui.Context, now time.Time) {
	next := now.Truncate(time.Minute).Add(time.Minute)
	c.After(next.Sub(now) + 10*time.Millisecond)
}

func relativeTime(now, fetched time.Time) string {
	minutes := int(now.Sub(fetched).Minutes())
	if minutes < 1 {
		return "刚刚"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d 分钟前", minutes)
	}

	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%d 小时前", hours)
	}
	return fmt.Sprintf("%d 天前", hours/24)
}
