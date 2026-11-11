package gui

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

const (
	// Floating sidebar geometry, in DIPs.
	feedSidebarHotZoneWidth  = 20
	feedSidebarFloatingInset = 8
	feedSidebarFloatingWidth = 300
	feedSidebarHideDelay     = 200 * time.Millisecond
	feedSidebarPaneSaveDelay = 500 * time.Millisecond

	// feedSidebarTextSize matches the webview's 13px list titles.
	feedSidebarTextSize = 13

	// Icons are revalidated weekly, matching the webview cache.
	feedSidebarIconTTL = 7 * 24 * time.Hour
)

// nativeTextSize is MyGO's default text size before system text scaling.
var nativeTextSize = ui.LightTheme().FontSize

// FeedSidebarCallbacks connects the native sidebar to application actions.
type FeedSidebarCallbacks struct {
	Icon           func(context.Context, int64) (store.Icon, error)
	Select         func(int64)
	RefreshAll     func()
	RefreshFeed    func(int64)
	MarkAllRead    func(int64)
	Remove         func(parent *mygo.Window, feedID int64)
	OpenAddFeed    func(parent *mygo.Window)
	OpenSettings   func(parent *mygo.Window)
	SetCollapsed   func(collapsed bool)
	SetPanePercent func(percent float32)
}

// sidebarIcon is one decoded feed icon.
type sidebarIcon struct {
	source  ui.ImageSource
	feedURL string
	fetched time.Time
	missing bool
}

// sidebarOverlay tracks the floating sidebar's slide state.
type sidebarOverlay int

const (
	sidebarHidden sidebarOverlay = iota
	sidebarShowing
	sidebarVisible
	sidebarHiding
)

// FeedSidebar renders subscriptions and their actions natively.
type FeedSidebar struct {
	mu         sync.Mutex
	model      *appstatus.Model
	callbacks  FeedSidebarCallbacks
	invalidate func()
	window     *mygo.Window

	panePercent   float32
	paneSaveTimer *time.Timer

	search       string
	searchOpen   bool
	icons        map[int64]sidebarIcon
	loadingIcons map[int64]bool

	collapsed bool
	overlay   sidebarOverlay
	entered   bool
	hideTimer *time.Timer
}

// NewFeedSidebar creates the native subscription sidebar.
func NewFeedSidebar(model *appstatus.Model, callbacks FeedSidebarCallbacks) *FeedSidebar {
	snapshot := model.Snapshot()
	return &FeedSidebar{
		model:        model,
		callbacks:    callbacks,
		panePercent:  snapshot.SidebarPercent,
		icons:        make(map[int64]sidebarIcon),
		loadingIcons: make(map[int64]bool),
		collapsed:    snapshot.SidebarCollapsed,
	}
}

// Attach connects the sidebar to the main window for child surfaces.
func (s *FeedSidebar) Attach(window *mygo.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.window = window
}

// SetInvalidate installs the callback used after asynchronous updates.
func (s *FeedSidebar) SetInvalidate(invalidate func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalidate = invalidate
}

// SidebarTheme returns the theme used by the sidebar and its splitter.
func (s *FeedSidebar) SidebarTheme(c *ui.Context) *ui.Theme {
	snapshot := s.model.Snapshot()
	return feedSidebarTheme(c, snapshot.Theme, snapshot.FontSize)
}

// feedSidebarTheme sizes the native sidebar text like the webview's list titles.
func feedSidebarTheme(c *ui.Context, appearance, fontSize string) *ui.Theme {
	theme := nativeTheme(c, appearance, fontSize)
	if nativeTextSize > 0 {
		theme.FontSize *= feedSidebarTextSize / nativeTextSize
	}
	return theme
}

// SyncLayout copies the mirrored sidebar layout into the splitter.
func (s *FeedSidebar) SyncLayout() {
	snapshot := s.model.Snapshot()
	s.mu.Lock()
	s.panePercent = snapshot.SidebarPercent
	s.collapsed = snapshot.SidebarCollapsed
	invalidate := s.invalidate
	s.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

// splitPercent exposes the docked sidebar's splitter position.
func (s *FeedSidebar) splitPercent() *float32 {
	return &s.panePercent
}

// isCollapsed reports whether the reader currently uses the floating sidebar.
func (s *FeedSidebar) isCollapsed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.collapsed
}

// savePanePercent schedules a pane-width preference update after dragging.
func (s *FeedSidebar) savePanePercent() {
	s.mu.Lock()
	percent := s.panePercent
	if s.paneSaveTimer != nil {
		s.paneSaveTimer.Stop()
	}
	callbacks := s.callbacks
	invalidate := s.invalidate
	s.paneSaveTimer = time.AfterFunc(feedSidebarPaneSaveDelay, func() {
		s.mu.Lock()
		s.paneSaveTimer = nil
		s.mu.Unlock()
		callbacks.SetPanePercent(percent)
	})
	s.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

// View builds the docked sidebar content.
func (s *FeedSidebar) View(c *ui.Context) {
	snapshot := s.model.Snapshot()
	theme := feedSidebarTheme(c, snapshot.Theme, snapshot.FontSize)
	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Background(theme.Background).Gap(8).Padding(10, 10).Children(func() {
			s.drawHeader(c, theme, snapshot)
			s.drawFeeds(c, theme, snapshot)
		})
	})
}

// drawOverlay draws the collapsed sidebar's hover zone and floating panel.
func (s *FeedSidebar) drawOverlay(c *ui.Context) {
	snapshot := s.model.Snapshot()
	theme := feedSidebarTheme(c, snapshot.Theme, snapshot.FontSize)

	hot := ui.Box(c).Absolute().Top(0).Left(0).
		Bottom(statusBarHeight(snapshot.FontSize)).
		Width(feedSidebarHotZoneWidth).Background(ui.Transparent)
	hotHovered := hot.Hovered()
	s.beginShow(hotHovered)

	if s.overlayState() == sidebarHidden {
		return
	}

	panel := ui.Column(c).Key("floating-sidebar").
		Absolute().
		Top(feedSidebarFloatingInset).
		Bottom(statusBarHeight(snapshot.FontSize)+feedSidebarFloatingInset).
		Left(feedSidebarFloatingInset).
		Width(feedSidebarFloatingWidth).
		Radius(12).
		Background(theme.Background).
		Border(1, theme.Border).
		Clip().
		Shadow(0, 8, 24, 0, ui.RGBA(0, 0, 0, 0.18)).
		Children(func() {
			s.View(c)
		})
	panelHovered := panel.Hovered()
	panel.IgnoreInput(s.overlayState() == sidebarHiding)

	nextFrame := false
	pose := ui.Pose{}
	switch s.overlayState() {
	case sidebarShowing:
		if !s.entered {
			pose = ui.Pose{XPercent: -104, Fade: 1}
			s.setEntered(true)
			nextFrame = true
		} else {
			s.setOverlayState(sidebarVisible)
		}
	case sidebarVisible:
		if !hotHovered && !panelHovered {
			s.setOverlayState(sidebarHiding)
			s.scheduleHide()
		}
	case sidebarHiding:
		if s.entered {
			s.setEntered(false)
			nextFrame = true
		} else {
			pose = ui.Pose{XPercent: -104, Fade: 1}
		}
	default:
		pose = ui.Pose{XPercent: -104, Fade: 1}
	}
	panel.AnimatePose(pose, ui.Snappy)
	if nextFrame && s.invalidate != nil {
		s.invalidate()
	}
}

func (s *FeedSidebar) drawHeader(c *ui.Context, theme *ui.Theme, snapshot appstatus.Snapshot) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Grow(1).AlignItems(ui.Center).Gap(6).Children(func() {
			ui.Text(c, "订阅源").FontWeight(600).Shrink(1).Ellipsis("…")
			ui.Badge(c, formatInt(snapshot.FeedCount)).Shrink(0)
		})

		var collapseIcon *ui.SVG
		collapseLabel := "恢复三栏布局"
		collapseAction := func() { s.runAsync(func() { s.callbacks.SetCollapsed(false) }) }
		if snapshot.SidebarCollapsed {
			collapseIcon = feedSidebarChevronRight
		} else {
			collapseIcon = feedSidebarChevronLeft
			collapseLabel = "浮动显示订阅源"
			collapseAction = func() { s.runAsync(func() { s.callbacks.SetCollapsed(true) }) }
		}
		s.iconButton(c, theme, collapseIcon, collapseLabel, collapseAction)

		search := s.iconButton(c, theme, feedSidebarSearch, "搜索订阅源", func() {
			s.searchOpen = !s.searchOpen
		})
		ui.Popover(c, search, &s.searchOpen, func() {
			ui.TextInput(c, &s.search).Placeholder("按名称或地址过滤").Height(32).Width(256)
		})

		s.iconButton(c, theme, feedSidebarPlus, "添加订阅源", func() {
			s.runAsync(func() { s.callbacks.OpenAddFeed(s.window) })
		})
		s.iconButton(c, theme, feedSidebarRefresh, "刷新全部", func() {
			s.runAsync(s.callbacks.RefreshAll)
		}).Disabled(snapshot.Sync.Active)
		s.iconButton(c, theme, feedSidebarSettings, "设置", func() {
			s.runAsync(func() { s.callbacks.OpenSettings(s.window) })
		})
	})
}

func (s *FeedSidebar) iconButton(
	c *ui.Context,
	theme *ui.Theme,
	icon *ui.SVG,
	label string,
	action func(),
) ui.Element {
	var button ui.Element
	button = ui.ButtonBase(c).
		Size(28, 28).
		Radius(6).
		Cursor(ui.CursorPointer).
		Label(label).
		Tooltip(label).
		Children(func() {
			ui.Icon(c, icon).Size(16, 16)
		})

	switch {
	case button.Pressed():
		button.Background(theme.SurfacePressed)
	case button.Hovered():
		button.Background(theme.SurfaceHover)
	default:
		button.Background(ui.Transparent)
	}
	if button.Clicked() {
		action()
	}
	return button
}

func (s *FeedSidebar) drawFeeds(c *ui.Context, theme *ui.Theme, snapshot appstatus.Snapshot) {
	feeds := visibleSidebarFeeds(snapshot.Feeds, s.search)
	if len(feeds) == 0 {
		ui.Column(c).Grow(1).Center().Children(func() {
			if len(snapshot.Feeds) == 0 {
				ui.Text(c, "还没有订阅源").TextColor(theme.TextMuted)
				ui.Text(c, "点击上方的 + 添加一个 RSS、Atom 或 JSON Feed 地址。").TextColor(theme.TextMuted)
			} else {
				ui.Text(c, "没有匹配的订阅源").TextColor(theme.TextMuted)
			}
		})
		return
	}

	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Fill().Gap(2).Children(func() {
			for _, feed := range feeds {
				s.drawFeed(c, theme, snapshot, feed)
			}
		})
	})
}

func (s *FeedSidebar) drawFeed(
	c *ui.Context,
	theme *ui.Theme,
	snapshot appstatus.Snapshot,
	feed store.Feed,
) {
	selected := snapshot.CurrentFeedID == feed.ID
	row := ui.ButtonBase(c).
		Key("feed-"+formatFeedID(feed.ID)).
		Grow(1).
		Height(34).
		Padding(10, 8).
		Radius(6).
		Background(feedRowBackground(*theme, selected)).
		Children(func() {
			s.drawFeedIcon(c, theme, feed)
			title := ui.Text(c, feed.Title).Grow(1).PaddingStart(10)
			if selected {
				title = title.Bold()
			}
			title.Ellipsis("…")
			if snapshot.Sync.Active && snapshot.Sync.FeedID == feed.ID {
				ui.Spinner(c)
			}
			if feed.LastError != "" {
				ui.Box(c).Size(8, 8).Radius(4).Background(theme.Danger).
					Description(feed.LastError)
			}
			if feed.Unread > 0 {
				ui.Badge(c, unreadText(feed.Unread))
			}
		})
	if !selected && row.Hovered() {
		row.Background(theme.SurfaceHover)
	}
	if row.Clicked() {
		s.runAsync(func() { s.callbacks.Select(feed.ID) })
	}
	row.ContextMenu(func(menu *ui.Menu) {
		if menu.Item("打开").Chosen() {
			s.runAsync(func() { s.callbacks.Select(feed.ID) })
		}
		if menu.Item("刷新").Chosen() {
			s.runAsync(func() { s.callbacks.RefreshFeed(feed.ID) })
		}
		if menu.Item("标记全部已读").Chosen() {
			s.runAsync(func() { s.callbacks.MarkAllRead(feed.ID) })
		}
		if menu.Item("取消订阅").Chosen() {
			s.removeFeed(feed.ID)
		}
	})
}

func (s *FeedSidebar) drawFeedIcon(c *ui.Context, theme *ui.Theme, feed store.Feed) {
	icon, ok := s.cachedIcon(feed)
	switch {
	case ok && !icon.missing:
		ui.Image(c, icon.source).Size(16, 16)
	case ok && icon.missing:
		ui.Icon(c, feedSidebarRss).Size(16, 16).TextColor(theme.TextMuted)
	default:
		ui.Icon(c, feedSidebarRss).Size(16, 16).TextColor(theme.TextMuted)
		s.loadIcon(feed.ID)
	}
}

func (s *FeedSidebar) loadIcon(feedID int64) {
	feed, ok := s.feedByID(feedID)
	if !ok {
		return
	}
	s.mu.Lock()
	if s.loadingIcons[feedID] {
		s.mu.Unlock()
		return
	}
	s.loadingIcons[feedID] = true
	callbacks := s.callbacks
	invalidate := s.invalidate
	s.mu.Unlock()

	go func() {
		icon, err := callbacks.Icon(context.Background(), feedID)
		decoded := sidebarIcon{feedURL: feed.IconURL, missing: err != nil || len(icon.Data) == 0}
		if err == nil && len(icon.Data) > 0 {
			if source, decodeErr := ui.DecodeImage(icon.Data); decodeErr == nil {
				decoded = sidebarIcon{source: source, feedURL: feed.IconURL, fetched: time.Now()}
			}
		}
		s.mu.Lock()
		s.icons[feedID] = decoded
		delete(s.loadingIcons, feedID)
		s.mu.Unlock()
		if invalidate != nil {
			invalidate()
		}
	}()
}

func (s *FeedSidebar) cachedIcon(feed store.Feed) (sidebarIcon, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	icon, ok := s.icons[feed.ID]
	if !ok || icon.feedURL != feed.IconURL {
		return sidebarIcon{}, false
	}
	if icon.missing {
		return icon, true
	}
	return icon, time.Since(icon.fetched) < feedSidebarIconTTL
}

func (s *FeedSidebar) feedByID(feedID int64) (store.Feed, bool) {
	for _, feed := range s.model.Snapshot().Feeds {
		if feed.ID == feedID {
			return feed, true
		}
	}
	return store.Feed{}, false
}

func (s *FeedSidebar) removeFeed(feedID int64) {
	s.mu.Lock()
	window := s.window
	callbacks := s.callbacks
	s.mu.Unlock()
	s.runAsync(func() { callbacks.Remove(window, feedID) })
}

func (s *FeedSidebar) runAsync(operation func()) {
	if operation == nil {
		return
	}
	go operation()
}

func (s *FeedSidebar) beginShow(hovered bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.collapsed {
		s.overlay = sidebarHidden
		s.entered = false
		return
	}
	if hovered && s.overlay == sidebarHidden {
		s.overlay = sidebarShowing
		s.entered = false
		if s.invalidate != nil {
			s.invalidate()
		}
	}
}

func (s *FeedSidebar) overlayState() sidebarOverlay {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.overlay
}

func (s *FeedSidebar) setOverlayState(state sidebarOverlay) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overlay = state
}

func (s *FeedSidebar) setEntered(entered bool) {
	s.mu.Lock()
	s.entered = entered
	s.mu.Unlock()
}

func (s *FeedSidebar) scheduleHide() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hideTimer != nil {
		s.hideTimer.Stop()
	}
	s.hideTimer = time.AfterFunc(feedSidebarHideDelay, func() {
		s.mu.Lock()
		s.overlay = sidebarHidden
		s.entered = false
		s.hideTimer = nil
		invalidate := s.invalidate
		s.mu.Unlock()
		if invalidate != nil {
			invalidate()
		}
	})
}

func visibleSidebarFeeds(feeds []store.Feed, filter string) []store.Feed {
	needle := strings.ToLower(strings.TrimSpace(filter))
	if needle == "" {
		return feeds
	}
	return slices.Clone(slices.DeleteFunc(feeds, func(feed store.Feed) bool {
		return !strings.Contains(strings.ToLower(feed.Title), needle) &&
			!strings.Contains(strings.ToLower(feed.URL), needle)
	}))
}

func feedRowBackground(theme ui.Theme, selected bool) ui.Color {
	if selected {
		return theme.Accent.Alpha(0.22)
	}
	return theme.Background
}

func unreadText(unread int) string {
	if unread > 99 {
		return "99+"
	}
	return formatInt(unread)
}

func formatFeedID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}
