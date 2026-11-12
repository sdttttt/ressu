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
	"runtime/pprof"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/logging"
	"github.com/sdttttt/ressu/core/media"
	"github.com/sdttttt/ressu/core/service"
	appstatus "github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
	"github.com/sdttttt/ressu/internal/action"
	"github.com/sdttttt/ressu/internal/buildinfo"
	"github.com/sdttttt/ressu/internal/gui"
	appabout "github.com/sdttttt/ressu/internal/gui/about"
	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	apppostheader "github.com/sdttttt/ressu/internal/gui/postheader"
	apppostlist "github.com/sdttttt/ressu/internal/gui/postlist"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
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
	stopProfile := startCPUProfile()
	defer stopProfile()
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
	listPercent := float32(30)
	if len(prefs.PaneSizes) == 2 {
		listPercent = float32(prefs.PaneSizes[1])
	}
	statusModel.SetListPercent(listPercent)
	statusModel.SetPostSort(prefs.PostSort)
	hub := appevent.NewHub(statusModel, webendpoint.AppEvent.Broadcast)
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
	notifications := appdialog.NewNotifications(func() (string, string) {
		snapshot := statusModel.Snapshot()
		return snapshot.Theme, snapshot.FontSize
	})
	confirmations := appdialog.NewConfirmations(func() (string, string) {
		prefs := st.Prefs()
		return prefs.Theme, prefs.FontSize
	})
	help := appdialog.NewHelp(func() (string, string) {
		prefs := st.Prefs()
		return prefs.Theme, prefs.FontSize
	})
	nativeUI := gui.NewNativeUI(notifications, confirmations)
	actions := action.New(feeds, posts, settings, status, hub, nativeUI)
	sidebar := appsidebar.New(statusModel, gui.SidebarCallbacks(actions))
	postList := apppostlist.New(statusModel, gui.PostListCallbacks(actions))
	postHeader := apppostheader.New(statusModel, gui.PostHeaderCallbacks(actions))
	nativeUI.SetHelp(help)
	nativeUI.SetSidebar(sidebar)
	nativeUI.SetPostList(postList)
	mygo.BindAs("Feeds", webendpoint.NewWebEndpoint(actions))
	mygo.Bind(
		webendpoint.NewPosts(actions),
		webendpoint.NewSettings(actions),
		webendpoint.NewSnapshot(actions),
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
		setMenu(actions)
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
		}, foreground, app.IsActive, about, notifications, confirmations, help, sidebar, postHeader, postList); err != nil {
			fatal("main window creation failed", err)
		}
	})

	if err := app.Run(); err != nil {
		fatal("application run failed", err)
	}
	slog.Info("application stopped", "operation", "app.stop")
}

// newAboutInfo gathers the metadata shown by the native About window.
func newAboutInfo(version string) appabout.Info {
	buildDate := "未知"
	if date := buildinfo.Date(); !date.IsZero() {
		buildDate = date.Format("2006-01-02")
	}
	return appabout.Info{
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

// startCPUProfile records a CPU profile for PGO when RESSU_CPUPROFILE names
// a writable file, and returns the function that stops and flushes it.
func startCPUProfile() func() {
	path := os.Getenv("RESSU_CPUPROFILE")
	if path == "" {
		return func() {}
	}
	file, err := os.Create(path)
	if err != nil {
		slog.Warn("cpu profile disabled", "operation", "app.profile", "path", path, "error", err)
		return func() {}
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		slog.Warn("cpu profile disabled", "operation", "app.profile", "path", path, "error", err)
		_ = file.Close()
		return func() {}
	}
	slog.Info("cpu profile recording", "operation", "app.profile", "path", path)
	return func() {
		pprof.StopCPUProfile()
		if err := file.Close(); err != nil {
			slog.Warn("cpu profile flush failed", "operation", "app.profile", "path", path, "error", err)
		}
	}
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
	about appabout.Info,
	notifications *appdialog.Notifications,
	confirmations *appdialog.Confirmations,
	help *appdialog.Help,
	sidebar *appsidebar.Sidebar,
	postHeader *apppostheader.Header,
	postList *apppostlist.List,
) (*mygo.Window, error) {
	var win *mygo.Window
	view := gui.NewAppView(statusModel, func() {
		info := about
		snapshot := statusModel.Snapshot()
		info.Theme = snapshot.Theme
		info.FontSize = snapshot.FontSize
		appabout.Open(win, info)
	}, notifications, confirmations, help, sidebar, postHeader, postList)
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
	help.Attach(win)
	sidebar.Attach(win)
	postHeader.Attach(win)
	sidebar.SetInvalidate(func() {
		if !win.IsDestroyed() {
			win.Invalidate()
		}
	})
	postList.Attach(win)
	postList.SetInvalidate(func() {
		if !win.IsDestroyed() {
			win.Invalidate()
		}
	})
	statusModel.SetChanged(func() {
		sidebar.SyncLayout()
		postList.SyncLayout()
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
