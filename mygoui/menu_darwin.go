//go:build darwin

package mygoui

import (
	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/app"
)

// setMenu installs the system menu bar, which macOS owns.
func setMenu(actions *app.App) {
	mygo.App.SetMenu(newMenu(actions))
}

// newMenu builds the menu bar. Its items call application actions directly.
func newMenu(actions *app.App) *mygo.Menu {
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
