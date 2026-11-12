package gui

import (
	"context"
	"errors"
	"log/slog"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/internal/action"
	apppostheader "github.com/sdttttt/ressu/internal/gui/postheader"
	apppostlist "github.com/sdttttt/ressu/internal/gui/postlist"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
)

// SidebarCallbacks wires the native subscription sidebar to application
// actions.
func SidebarCallbacks(a *action.AppActions) appsidebar.Callbacks {
	return appsidebar.Callbacks{
		Icon:        a.FeedIcon,
		Select:      func(feedID int64) { a.SelectFeed(context.Background(), feedID) },
		RefreshAll:  func() { refreshAllWithNotice(a) },
		RefreshFeed: func(feedID int64) { refreshFeedWithNotice(a, feedID) },
		MarkAllRead: func(feedID int64) { markAllReadWithNotice(a, feedID) },
		Remove: func(parent *mygo.Window, feedID int64) {
			if err := a.ConfirmRemoveFeed(context.Background(), action.NewWindowHandle(parent), feedID); err != nil {
				slog.Warn("native feed removal failed",
					"operation", "gui.sidebar.remove", "feedID", feedID, "error", err)
				a.Notify("error", "取消订阅失败")
			}
		},
		OpenAddFeed:    func(parent *mygo.Window) { a.OpenAddFeed(action.NewWindowHandle(parent)) },
		OpenSettings:   func(parent *mygo.Window) { a.OpenSettings(action.NewWindowHandle(parent)) },
		SetCollapsed:   a.SetSidebarCollapsed,
		SetPanePercent: a.SetSidebarPercent,
	}
}

// PostListCallbacks wires the native article list to application actions.
func PostListCallbacks(a *action.AppActions) apppostlist.Callbacks {
	return apppostlist.Callbacks{
		List:           a.ListPosts,
		Select:         a.SelectPost,
		ToggleStar:     a.TogglePostStar,
		MarkAllRead:    func(feedID int64) { markAllReadWithNotice(a, feedID) },
		SetPanePercent: a.SetListPercent,
		Notify:         a.Notify,
	}
}

// PostHeaderCallbacks wires the native article header to application actions.
func PostHeaderCallbacks(a *action.AppActions) apppostheader.Callbacks {
	return apppostheader.Callbacks{
		ToggleStar:      a.TogglePostStar,
		MarkRead:        a.MarkPostRead,
		OpenLink:        a.OpenLink,
		RequestSnapshot: a.RequestSnapshot,
		Notify:          a.Notify,
	}
}

func refreshAllWithNotice(a *action.AppActions) {
	if _, err := a.RefreshAll(context.Background()); err != nil {
		if errors.Is(err, service.ErrRefreshing) {
			return
		}
		slog.Warn("native refresh all failed",
			"operation", "gui.sidebar.refreshAll", "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func refreshFeedWithNotice(a *action.AppActions, feedID int64) {
	if _, err := a.RefreshFeed(context.Background(), feedID); err != nil {
		slog.Warn("native feed refresh failed",
			"operation", "gui.sidebar.refreshFeed", "feedID", feedID, "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func markAllReadWithNotice(a *action.AppActions, feedID int64) {
	if _, err := a.MarkAllPostsRead(feedID); err != nil {
		slog.Warn("native mark all read failed",
			"operation", "gui.sidebar.markAllRead", "feedID", feedID, "error", err)
		a.Notify("error", "标记已读失败")
	}
}
