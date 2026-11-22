package mygoui

import (
	"context"
	"errors"
	"log/slog"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/app"
	apppostheader "github.com/sdttttt/ressu/mygoui/postheader"
	apppostlist "github.com/sdttttt/ressu/mygoui/postlist"
	appsidebar "github.com/sdttttt/ressu/mygoui/sidebar"
)

// SidebarCallbacks wires the native subscription sidebar to application
// actions.
func SidebarCallbacks(a *app.App) appsidebar.Callbacks {
	return appsidebar.Callbacks{
		Icon:   a.FeedIcon,
		Select: func(feedID int64) { a.SelectFeed(context.Background(), feedID) },
		Move: func(draggedID int64, to int) {
			if err := a.MoveFeed(draggedID, to); err != nil {
				slog.Warn("native feed move failed",
					"operation", "mygoui.sidebar.move", "feedID", draggedID, "error", err)
				a.Notify("error", "调整订阅源顺序失败")
			}
		},
		RefreshAll:  func() { refreshAllWithNotice(a) },
		RefreshFeed: func(feedID int64) { refreshFeedWithNotice(a, feedID) },
		MarkAllRead: func(parent *mygo.Window, feedID int64) { markAllReadWithNotice(a, parent, feedID) },
		Remove: func(parent *mygo.Window, feedID int64) {
			if err := a.ConfirmRemoveFeed(context.Background(), app.NewWindowHandle(parent), feedID); err != nil {
				slog.Warn("native feed removal failed",
					"operation", "mygoui.sidebar.remove", "feedID", feedID, "error", err)
				a.Notify("error", "取消订阅失败")
			}
		},
		Edit: func(parent *mygo.Window, feedID int64) {
			a.OpenEditFeed(app.NewWindowHandle(parent), feedID)
		},
		OpenAddFeed:    func(parent *mygo.Window) { a.OpenAddFeed(app.NewWindowHandle(parent)) },
		OpenSettings:   func(parent *mygo.Window) { a.OpenSettings(app.NewWindowHandle(parent)) },
		SetCollapsed:   a.SetSidebarCollapsed,
		SetPanePercent: a.SetSidebarPercent,
	}
}

// PostListCallbacks wires the native article list to application actions.
func PostListCallbacks(a *app.App) apppostlist.Callbacks {
	return apppostlist.Callbacks{
		List:           a.ListPosts,
		Select:         a.SelectPost,
		ToggleRead:     a.TogglePostRead,
		ToggleStar:     a.TogglePostStar,
		MarkAllRead:    func(parent *mygo.Window, feedID int64) { markAllReadWithNotice(a, parent, feedID) },
		OpenLink:       a.OpenLink,
		CopyLink:       func(link string) { copyLinkToClipboard(a, link) },
		SetPanePercent: a.SetListPercent,
		Notify:         a.Notify,
	}
}

// PostHeaderCallbacks wires the native article header to application actions.
func PostHeaderCallbacks(a *app.App) apppostheader.Callbacks {
	return apppostheader.Callbacks{
		ToggleStar: a.TogglePostStar,
		MarkRead:   a.MarkPostRead,
		OpenLink:   a.OpenLink,
		CopyLink:   func(link string) { copyLinkToClipboard(a, link) },
		ExtractFullText: func(feedID int64, guid string) error {
			return a.ExtractPostFullText(context.Background(), feedID, guid)
		},
		RequestSnapshot: func(parent *mygo.Window) {
			a.RequestSnapshot(app.NewWindowHandle(parent))
		},
		Notify: a.Notify,
	}
}

// copyLinkToClipboard writes link to the system clipboard and confirms with
// a toast.
func copyLinkToClipboard(a *app.App, link string) {
	mygo.Clipboard.WriteText(link)
	a.Notify("success", "已复制链接")
}

func refreshAllWithNotice(a *app.App) {
	if _, err := a.RefreshAll(context.Background()); err != nil {
		if errors.Is(err, app.ErrRefreshing) {
			return
		}
		slog.Warn("native refresh all failed",
			"operation", "mygoui.sidebar.refreshAll", "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func refreshFeedWithNotice(a *app.App, feedID int64) {
	if _, err := a.RefreshFeed(context.Background(), feedID); err != nil {
		slog.Warn("native feed refresh failed",
			"operation", "mygoui.sidebar.refreshFeed", "feedID", feedID, "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func markAllReadWithNotice(a *app.App, parent *mygo.Window, feedID int64) {
	if _, err := a.ConfirmMarkAllRead(context.Background(), app.NewWindowHandle(parent), feedID); err != nil {
		slog.Warn("native mark all read failed",
			"operation", "mygoui.sidebar.markAllRead", "feedID", feedID, "error", err)
		a.Notify("error", "标记已读失败")
	}
}
