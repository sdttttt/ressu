// Ressu is a desktop RSS reader: the Go side fetches and parses feeds, keeps
// them in a bbolt database under the user data directory, and serves them to
// the Solid frontend through MyGo bindings.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/action"
	appevent "github.com/sdttttt/ressu/internal/event"
	"github.com/sdttttt/ressu/internal/gui"
	"github.com/sdttttt/ressu/internal/logging"
	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/service"
	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
	"github.com/sdttttt/ressu/internal/webendpoint"
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

	articleMedia, err := media.Open(filepath.Join(cacheDir, "articles"), media.Options{
		MaxBytes: int64(st.Prefs().ImageCacheLimit) << 20,
	})
	if err != nil {
		fatal("article media cache open failed", err)
	}
	defer articleMedia.Close()
	if err := mygo.Protocol.Handle(media.Scheme, articleMedia.Handler()); err != nil {
		fatal("media protocol registration failed", err)
	}

	prefs := st.Prefs()
	statusModel := appstatus.New(app.Version())
	statusModel.SetAppearance(prefs.Theme, prefs.FontSize)
	hub := appevent.NewHub(statusModel)
	feeds := service.NewFeedsWithHub(st, hub)
	hub.FeedsChanged(feeds.List())
	posts := service.NewPostsWithMediaAndHub(st, articleMedia, hub)
	settings := service.NewSettingsWithMedia(st, articleMedia)
	pageReady := make(chan struct{}, 1)
	status := service.NewStatus(hub, func() {
		select {
		case pageReady <- struct{}{}:
		default:
		}
	})
	actions := action.New(feeds, posts, settings, status, hub, gui.NativeUI{})
	mygo.BindAs("Feeds", webendpoint.NewWebEndpoint(actions))
	mygo.Bind(
		webendpoint.NewPosts(actions),
		webendpoint.NewSettings(actions),
		webendpoint.NewStatus(actions),
	)
	foreground := service.NewForegroundRefresher(func(ctx context.Context) {
		if _, err := feeds.RefreshStale(ctx); err != nil && !errors.Is(err, service.ErrRefreshing) {
			slog.Warn("refresh on foreground failed",
				"operation", "app.foregroundRefresh", "error", err)
		}
	})
	app.OnDidResignActive(foreground.ResignActive)
	app.OnDidBecomeActive(foreground.BecomeActive)

	app.WhenReady(func() {
		prefs := st.Prefs()
		gui.NativeUI{}.ApplyTheme(prefs.Theme)
		setMenu(hub, actions)
		if _, err := openWindow(statusModel, pageReady, func() {
			if !prefs.RefreshOnStart {
				return
			}
			go func() {
				if _, err := feeds.RefreshStale(context.Background()); err != nil {
					slog.Warn("refresh on start failed",
						"operation", "app.startRefresh", "error", err)
				}
			}()
		}, foreground, app.IsActive); err != nil {
			fatal("main window creation failed", err)
		}
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
// mounted, so startup shows a splash instead of a blank window. onReady runs
// once the window is shown.
func openWindow(statusModel *appstatus.Model, pageReady <-chan struct{}, onReady func(), foreground *service.ForegroundRefresher, isActive func() bool) (*mygo.Window, error) {
	view := gui.NewAppView(statusModel)
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:           "Ressu",
		Width:           1200,
		Height:          800,
		MinWidth:        900,
		MinHeight:       600,
		StateKey:        "main",
		Hidden:          true,
		BackgroundColor: "light-dark(#ffffff, #101012)",
		Content:         ui.View(view.View),
	})
	statusModel.SetChanged(func() {
		if !win.IsDestroyed() {
			win.Invalidate()
		}
	})

	var once sync.Once
	show := func() {
		if win.IsDestroyed() {
			return
		}
		once.Do(func() {
			win.Show()
			onReady()
		})
	}

	page, err := win.NewWebView(mygo.WebViewOptions{
		URL:             "/",
		BackgroundColor: "light-dark(#ffffff, #101012)",
	})
	if err != nil {
		win.Destroy()
		return nil, fmt.Errorf("create reader page: %w", err)
	}
	view.SetPage(page)
	win.Invalidate()

	timer := time.AfterFunc(5*time.Second, show)
	go func() {
		<-pageReady
		timer.Stop()
		show()
	}()
	win.OnBlur(func() { foreground.WindowBlurred(isActive()) })
	win.OnFocus(func() { foreground.WindowFocused(isActive()) })
	win.OnClosed(func() { timer.Stop() })
	return win, nil
}
