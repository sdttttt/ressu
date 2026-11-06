// Ressu is a desktop RSS reader: the Go side fetches and parses feeds, keeps
// them in a bbolt database under the user data directory, and serves them to
// the Solid frontend through MyGo bindings.
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/service"
	"github.com/sdttttt/ressu/internal/store"
)

func main() {
	app := mygo.App
	app.SetName("Ressu")

	dir, err := app.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	st, err := store.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	feeds := service.NewFeeds(st)
	mygo.Bind(feeds, service.NewPosts(st), service.NewSettings(st))

	app.WhenReady(func() {
		prefs := st.Prefs()
		service.ApplyTheme(prefs.Theme)
		app.SetMenu(newMenu())
		openWindow(func() {
			if !prefs.RefreshOnStart {
				return
			}
			go func() {
				if _, err := feeds.RefreshAll(context.Background()); err != nil {
					log.Printf("refresh on start: %v", err)
				}
			}()
		})
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// openWindow creates the main window hidden and shows it once the page has
// painted, so startup shows a splash instead of a blank window. onReady runs
// once the window is shown.
func openWindow(onReady func()) {
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:           "Ressu",
		URL:             "/",
		Width:           1200,
		Height:          800,
		MinWidth:        900,
		MinHeight:       600,
		StateKey:        "main",
		Hidden:          true,
		BackgroundColor: "light-dark(#ffffff, #101012)",
	})

	var once sync.Once
	show := func() {
		once.Do(func() {
			win.Show()
			onReady()
		})
	}

	win.OnReadyToShow(show)
	// A page that never becomes ready must not leave the app windowless.
	timer := time.AfterFunc(5*time.Second, show)
	win.OnClosed(func() { timer.Stop() })
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
