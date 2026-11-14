//go:build darwin

package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/app"
	"github.com/sdttttt/ressu/core/service"
)

// setMenu installs the system menu bar, which macOS owns.
func setMenu(actions *app.AppActions) {
	mygo.App.SetMenu(newMenu(actions))
}

// newMenu builds the menu bar. Its items call application actions directly.
func newMenu(actions *app.AppActions) *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "Add Feed…", Accelerator: "CmdOrCtrl+N", Click: func(_ *mygo.MenuItem, window *mygo.Window) {
				actions.OpenAddFeed(app.NewWindowHandle(window))
			}},
			{Label: "Refresh All", Accelerator: "CmdOrCtrl+R", Click: func(_ *mygo.MenuItem, _ *mygo.Window) {
				refreshAllWithNotice(actions)
			}},
			mygo.Separator(),
			{Label: "Settings…", Accelerator: "CmdOrCtrl+,", Click: func(_ *mygo.MenuItem, window *mygo.Window) {
				actions.OpenSettings(app.NewWindowHandle(window))
			}},
			mygo.Separator(),
			{Role: mygo.RoleClose},
		}},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	})
}

func refreshAllWithNotice(actions *app.AppActions) {
	if _, err := actions.RefreshAll(context.Background()); err != nil {
		if errors.Is(err, service.ErrRefreshing) {
			return
		}
		slog.Warn("menu refresh all failed",
			"operation", "app.menu.refreshAll", "error", err)
		actions.Notify("error", "刷新订阅源失败")
	}
}
