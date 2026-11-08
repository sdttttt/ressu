// Ressu is a desktop RSS reader: the Go side fetches and parses feeds, keeps
// them in a bbolt database under the user data directory, and serves them to
// the Solid frontend through MyGo bindings.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/logging"
	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/service"
	"github.com/sdttttt/ressu/internal/store"
)

func main() {
	app := mygo.App
	app.SetName("Ressu")
	dir, err := app.Path(mygo.PathUserData)
	if err != nil {
		fatal("user data path failed", err)
	}
	logFile, err := logging.Open(dir, logging.Options{})
	if err != nil {
		slog.Warn("log file disabled", "operation", "logging.open", "error", err)
		logFile = nil
	} else {
		defer logFile.Close()
		slog.SetDefault(logFile.Logger())
	}
	cacheDir, err := app.Path(mygo.PathCache)
	if err != nil {
		fatal("cache path failed", err)
	}
	st, err := store.Open(dir)
	if err != nil {
		fatal("database open failed", err)
	}
	defer st.Close()
	slog.Info("application starting", "operation", "app.start")

	articleMedia, err := media.Open(filepath.Join(cacheDir, "articles"))
	if err != nil {
		fatal("article media cache open failed", err)
	}
	defer articleMedia.Close()
	if err := mygo.Protocol.Handle(media.Scheme, articleMedia.Handler()); err != nil {
		fatal("media protocol registration failed", err)
	}

	feeds := service.NewFeeds(st)
	mygo.Bind(
		feeds,
		service.NewPostsWithMedia(st, articleMedia),
		service.NewSettings(st),
	)
	foreground := service.NewForegroundRefresher(func(ctx context.Context) {
		if _, err := feeds.RefreshAll(ctx); err != nil && !errors.Is(err, service.ErrRefreshing) {
			slog.Warn("refresh on foreground failed",
				"operation", "app.foregroundRefresh", "error", err)
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
					slog.Warn("refresh on start failed",
						"operation", "app.startRefresh", "error", err)
				}
			}()
		}, foreground, app.IsActive)
	})

	if err := app.Run(); err != nil {
		fatal("application run failed", err)
	}
	slog.Info("application stopped", "operation", "app.stop")
}

// fatal records a startup error and exits before the application becomes usable.
func fatal(message string, err error) {
	slog.Error(message, "operation", "app.startup", "error", err)
	os.Exit(1)
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
