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
	"github.com/sdttttt/ressu/internal/buildinfo"
	appevent "github.com/sdttttt/ressu/internal/event"
	"github.com/sdttttt/ressu/internal/gui"
	"github.com/sdttttt/ressu/internal/logging"
	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/service"
	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
	"github.com/sdttttt/ressu/internal/webendpoint"
)

const (
	appDeveloper = "sdttttt"
	appEmail     = "sendaihazuki@foxmail.com"
)

func main() {
	app := mygo.App
	app.SetName("Ressu")
	if os.Getenv("MYGO_GENERATE") != "" {
		generateDir := filepath.Join(os.TempDir(), "ressu-generate")
		if err := os.MkdirAll(generateDir, 0o755); err != nil {
			fatal("generate data path failed", err)
		}
		app.SetPath(mygo.PathUserData, generateDir)
		app.SetPath(mygo.PathCache, filepath.Join(generateDir, "cache"))
		app.SetPath(mygo.PathLogs, filepath.Join(generateDir, "logs"))
	}
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
	sidebarPercent := float32(18)
	if len(prefs.PaneSizes) == 2 {
		sidebarPercent = float32(prefs.PaneSizes[0])
	}
	statusModel.SetLayout(prefs.SidebarCollapsed, sidebarPercent)
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
	notifications := gui.NewNotifications(func() (string, string) {
		snapshot := statusModel.Snapshot()
		return snapshot.Theme, snapshot.FontSize
	})
	confirmations := gui.NewConfirmations(func() store.Prefs {
		return st.Prefs()
	})
	nativeUI := gui.NewNativeUI(notifications, confirmations)
	actions := action.New(feeds, posts, settings, status, hub, nativeUI)
	sidebar := gui.NewFeedSidebar(statusModel, actions.SidebarCallbacks())
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
		nativeUI.ApplyTheme(prefs.Theme)
		setMenu(hub, actions)
		about := newAboutInfo(app.Version())
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
		}, foreground, app.IsActive, about, notifications, confirmations, sidebar); err != nil {
			fatal("main window creation failed", err)
		}
	})

	if err := app.Run(); err != nil {
		fatal("application run failed", err)
	}
	slog.Info("application stopped", "operation", "app.stop")
}

// newAboutInfo gathers the metadata shown by the native About window.
func newAboutInfo(version string) gui.AboutInfo {
	buildDate := "未知"
	if date := buildinfo.Date(); !date.IsZero() {
		buildDate = date.Format("2006-01-02")
	}
	return gui.AboutInfo{
		Version:   version,
		BuildDate: buildDate,
		Developer: appDeveloper,
		Email:     appEmail,
	}
}

// fatal records a startup error and exits before the application becomes usable.
func fatal(message string, err error) {
	slog.Error(message, "operation", "app.startup", "error", err)
	os.Exit(1)
}

// openWindow creates the main window hidden and shows it once the page has
// mounted, so startup shows a splash instead of a blank window. onReady runs
// once the window is shown.
func openWindow(
	statusModel *appstatus.Model,
	pageReady <-chan struct{},
	onReady func(),
	foreground *service.ForegroundRefresher,
	isActive func() bool,
	about gui.AboutInfo,
	notifications *gui.Notifications,
	confirmations *gui.Confirmations,
	sidebar *gui.FeedSidebar,
) (*mygo.Window, error) {
	var win *mygo.Window
	view := gui.NewAppView(statusModel, func() {
		info := about
		snapshot := statusModel.Snapshot()
		info.Theme = snapshot.Theme
		info.FontSize = snapshot.FontSize
		gui.OpenAboutWindow(win, info)
	}, notifications, confirmations, sidebar)
	win = mygo.NewWindow(mygo.WindowOptions{
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
	notifications.Attach(win)
	confirmations.Attach(win)
	sidebar.Attach(win)
	sidebar.SetInvalidate(func() {
		if !win.IsDestroyed() {
			win.Invalidate()
		}
	})
	statusModel.SetChanged(func() {
		sidebar.SyncLayout()
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
