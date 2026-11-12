//go:build darwin

package main

import (
	"github.com/egoist/mygo"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/internal/action"
)

// setMenu installs the system menu bar, which macOS owns.
func setMenu(hub *appevent.Hub, actions *action.AppActions) {
	mygo.App.SetMenu(newMenu(hub, actions))
}

// newMenu builds the menu bar. Its items send commands to the page, which owns
// the dialogs they open, except Settings, which opens natively.
func newMenu(hub *appevent.Hub, actions *action.AppActions) *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "Add Feed…", Accelerator: "CmdOrCtrl+N", Click: sendCommand(hub, service.CommandAddFeed)},
			{Label: "Refresh All", Accelerator: "CmdOrCtrl+R", Click: sendCommand(hub, service.CommandRefresh)},
			mygo.Separator(),
			{Label: "Settings…", Accelerator: "CmdOrCtrl+,", Click: func(_ *mygo.MenuItem, window *mygo.Window) {
				actions.OpenSettings(action.NewWindowHandle(window))
			}},
			mygo.Separator(),
			{Role: mygo.RoleClose},
		}},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	})
}

func sendCommand(hub *appevent.Hub, name string) func(*mygo.MenuItem, *mygo.Window) {
	return func(*mygo.MenuItem, *mygo.Window) {
		hub.UICommand(name)
	}
}
