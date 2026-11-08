// Ressu is a desktop RSS reader: the Go side fetches and parses feeds, keeps
// them in a bbolt database under the user data directory, and serves them to
// the Solid frontend through MyGo bindings.
package main

import (
	"context"
	"errors"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/media"
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
	cacheDir, err := app.Path(mygo.PathCache)
	if err != nil {
		log.Fatal(err)
	}
	st, err := store.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	articleMedia, err := media.Open(filepath.Join(cacheDir, "articles"))
	if err != nil {
		log.Fatal(err)
	}
	defer articleMedia.Close()
	if err := mygo.Protocol.Handle(media.Scheme, articleMedia.Handler()); err != nil {
		log.Fatal(err)
	}

	feeds := service.NewFeeds(st)
	mygo.Bind(feeds, service.NewPostsWithMedia(st, articleMedia), service.NewSettings(st))
	foreground := service.NewForegroundRefresher(func(ctx context.Context) {
		if _, err := feeds.RefreshAll(ctx); err != nil && !errors.Is(err, service.ErrRefreshing) {
			log.Printf("refresh on foreground: %v", err)
		}
	})
	app.OnDidResignActive(foreground.ResignActive)
	app.OnDidBecomeActive(foreground.BecomeActive)

	app.WhenReady(func() {
		prefs := st.Prefs()
		service.ApplyTheme(prefs.Theme)
		setMenu()
		openWindow(func() {
			if !prefs.RefreshOnStart {
				return
			}
			go func() {
				if _, err := feeds.RefreshAll(context.Background()); err != nil {
					log.Printf("refresh on start: %v", err)
				}
			}()
		}, foreground, app.IsActive)
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// openWindow creates the main window hidden and shows it once the page has
// painted, so startup shows a splash instead of a blank window. onReady runs
// once the window is shown.
func openWindow(onReady func(), foreground *service.ForegroundRefresher, isActive func() bool) {
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
	win.OnBlur(func() { foreground.WindowBlurred(isActive()) })
	win.OnFocus(func() { foreground.WindowFocused(isActive()) })
	// A page that never becomes ready must not leave the app windowless.
	timer := time.AfterFunc(5*time.Second, show)
	win.OnClosed(func() { timer.Stop() })
}
