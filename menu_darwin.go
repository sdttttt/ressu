//go:build darwin

package main

import (
	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/service"
)

// setMenu installs the system menu bar, which macOS owns.
func setMenu() {
	mygo.App.SetMenu(newMenu())
}

// newMenu builds the menu bar. Its items send commands to the page, which owns
// the dialogs they open.
func newMenu() *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "Add Feed…", Accelerator: "CmdOrCtrl+N", Click: sendCommand(service.CommandAddFeed)},
			{Label: "Refresh All", Accelerator: "CmdOrCtrl+R", Click: sendCommand(service.CommandRefresh)},
			mygo.Separator(),
			{Label: "Settings…", Accelerator: "CmdOrCtrl+,", Click: sendCommand(service.CommandSettings)},
			mygo.Separator(),
			{Role: mygo.RoleClose},
		}},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	})
}

func sendCommand(name string) func(*mygo.MenuItem, *mygo.Window) {
	return func(*mygo.MenuItem, *mygo.Window) {
		_ = service.Command.Broadcast(name)
	}
}
