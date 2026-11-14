// Package sidebar renders the native subscription list and floating overlay.
package sidebar

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
	appicon "github.com/sdttttt/ressu/mygoui/icon"
	guitheme "github.com/sdttttt/ressu/mygoui/theme"
	"github.com/sdttttt/ressu/mygoui/widget"
)

const (
	aggregateFeedID = 0

	// Floating sidebar geometry, in DIPs.
	hotZoneWidth  = 20
	floatingInset = 8
	floatingWidth = 300
	hideDelay     = 200 * time.Millisecond
	paneSaveDelay = 500 * time.Millisecond

	// textSize matches the webview's 13px list titles.
	textSize = 13

	// feedRowHeight and feedRowGap size one subscription row; the selection
	// pill slides along their pitch.
	feedRowHeight = 34
	feedRowGap    = 2

	// Icons are revalidated weekly, matching the webview cache.
	iconTTL = 7 * 24 * time.Hour
)

// nativeTextSize is MyGO's default text size before system text scaling.
var nativeTextSize = ui.LightTheme().FontSize

// selectionSlide paces the selected feed's highlight. It is quicker than
// the stock Snappy spring so keyboard runs keep up.
var selectionSlide = ui.Animation{
	Spring: ui.SpringOptions{Duration: 400 * time.Millisecond, Bounce: 0.15},
}

// poseSnap lands a pose at once, for the reduce-motion preference: a timed
// animation of a millisecond instead of the Snappy spring.
var poseSnap = ui.Animation{Duration: time.Millisecond}

// Callbacks connects the native sidebar to application actions.
type Callbacks struct {
	Icon           func(context.Context, int64) (app.Icon, error)
	Select         func(int64)
	RefreshAll     func()
	RefreshFeed    func(int64)
	MarkAllRead    func(int64)
	Remove         func(parent *mygo.Window, feedID int64)
	Edit           func(parent *mygo.Window, feedID int64)
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

// Sidebar renders subscriptions and their actions natively.
type Sidebar struct {
	mu         sync.Mutex
	model      *app.Model
	callbacks  Callbacks
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

// New creates the native subscription sidebar.
func New(model *app.Model, callbacks Callbacks) *Sidebar {
	snapshot := model.Snapshot()
	return &Sidebar{
		model:        model,
		callbacks:    callbacks,
		panePercent:  snapshot.SidebarPercent,
		icons:        make(map[int64]sidebarIcon),
		loadingIcons: make(map[int64]bool),
		collapsed:    snapshot.SidebarCollapsed,
	}
}

// Attach connects the sidebar to the main window for child surfaces.
func (s *Sidebar) Attach(window *mygo.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.window = window
}

// SetInvalidate installs the callback used after asynchronous updates.
func (s *Sidebar) SetInvalidate(invalidate func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalidate = invalidate
}

// Theme returns the theme used by the sidebar and its splitter.
func (s *Sidebar) Theme(c *ui.Context) *ui.Theme {
	snapshot := s.model.Snapshot()
	return Theme(c, snapshot.Theme, snapshot.FontSize)
}

// Theme sizes the native sidebar text like the webview's list titles.
func Theme(c *ui.Context, appearance, fontSize string) *ui.Theme {
	theme := guitheme.New(c, appearance, fontSize)
	if nativeTextSize > 0 {
		theme.FontSize *= textSize / nativeTextSize
	}
	return theme
}

// SyncLayout copies the mirrored sidebar layout into the splitter.
func (s *Sidebar) SyncLayout() {
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

// SplitPercent exposes the docked sidebar's splitter position.
func (s *Sidebar) SplitPercent() *float32 {
	return &s.panePercent
}

// IsCollapsed reports whether the reader currently uses the floating sidebar.
func (s *Sidebar) IsCollapsed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.collapsed
}

// SavePanePercent schedules a pane-width preference update after dragging.
func (s *Sidebar) SavePanePercent() {
	s.mu.Lock()
	percent := s.panePercent
	if s.paneSaveTimer != nil {
		s.paneSaveTimer.Stop()
	}
	callbacks := s.callbacks
	invalidate := s.invalidate
	s.paneSaveTimer = time.AfterFunc(paneSaveDelay, func() {
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

// SelectNeighbour selects the next or previous subscription without wrapping.
func (s *Sidebar) SelectNeighbour(delta int) {
	snapshot := s.model.Snapshot()
	if len(snapshot.Feeds) == 0 {
		return
	}

	index := slices.IndexFunc(snapshot.Feeds, func(feed app.Feed) bool {
		return feed.ID == snapshot.CurrentFeedID
	})
	if index == -1 {
		if delta < 0 {
			index = len(snapshot.Feeds) - 1
		} else {
			index = 0
		}
	} else {
		index = min(max(index+delta, 0), len(snapshot.Feeds)-1)
	}

	feedID := snapshot.Feeds[index].ID
	s.runAsync(func() { s.callbacks.Select(feedID) })
}

// View builds the docked sidebar content.
func (s *Sidebar) View(c *ui.Context) {
	snapshot := s.model.Snapshot()
	theme := Theme(c, snapshot.Theme, snapshot.FontSize)
	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Background(theme.Background).Gap(8).Padding(10, 10).Children(func() {
			s.drawHeader(c, theme, snapshot)
			s.drawFeeds(c, theme, snapshot)
		})
	})
}

// DrawOverlay draws the collapsed sidebar's hover zone and floating panel.
func (s *Sidebar) DrawOverlay(c *ui.Context) {
	snapshot := s.model.Snapshot()
	theme := Theme(c, snapshot.Theme, snapshot.FontSize)

	hot := ui.Box(c).Absolute().Top(0).Left(0).
		Bottom(guitheme.StatusBarHeight(snapshot.FontSize)).
		Width(hotZoneWidth).Background(ui.Transparent)
	hotHovered := hot.Hovered()
	s.beginShow(hotHovered)

	if s.overlayState() == sidebarHidden {
		return
	}

	panel := ui.Column(c).Key("floating-sidebar").
		Absolute().
		Top(floatingInset).
		Bottom(guitheme.StatusBarHeight(snapshot.FontSize)+floatingInset).
		Left(floatingInset).
		Width(floatingWidth).
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
	anim := ui.Snappy
	if snapshot.ReduceMotion {
		anim = poseSnap
	}
	panel.AnimatePose(pose, anim)
	if nextFrame && s.invalidate != nil {
		s.invalidate()
	}
}

func (s *Sidebar) drawHeader(c *ui.Context, theme *ui.Theme, snapshot app.Snapshot) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Grow(1).AlignItems(ui.Center).Gap(6).Children(func() {
			ui.Text(c, "订阅源").FontWeight(600).Shrink(1).Ellipsis("…")
			ui.Badge(c, formatInt(snapshot.FeedCount)).Shrink(0)
		})

		var collapseIcon *ui.SVG
		collapseLabel := "恢复三栏布局"
		collapseAction := func() { s.runAsync(func() { s.callbacks.SetCollapsed(false) }) }
		if snapshot.SidebarCollapsed {
			collapseIcon = appicon.ChevronRight
		} else {
			collapseIcon = appicon.ChevronLeft
			collapseLabel = "浮动显示订阅源"
			collapseAction = func() { s.runAsync(func() { s.callbacks.SetCollapsed(true) }) }
		}
		s.iconButton(c, theme, collapseIcon, collapseLabel, true, collapseAction)

		search := s.iconButton(c, theme, appicon.Search, "搜索订阅源", true, func() {
			s.searchOpen = !s.searchOpen
		})
		ui.Popover(c, search, &s.searchOpen, func() {
			ui.TextInput(c, &s.search).Placeholder("按名称或地址过滤").Height(32).Width(256)
		})

		s.iconButton(c, theme, appicon.Plus, "添加订阅源", true, func() {
			s.runAsync(func() { s.callbacks.OpenAddFeed(s.window) })
		})
		s.iconButton(c, theme, appicon.Refresh, "刷新全部", !snapshot.Sync.Active, func() {
			s.runAsync(s.callbacks.RefreshAll)
		})
		s.iconButton(c, theme, appicon.Settings, "设置", true, func() {
			s.runAsync(func() { s.callbacks.OpenSettings(s.window) })
		})
	})
}

func (s *Sidebar) iconButton(
	c *ui.Context,
	theme *ui.Theme,
	icon *ui.SVG,
	label string,
	enabled bool,
	action func(),
) ui.Element {
	return widget.IconButton(c, widget.ThemeIconButtonColors(theme), widget.IconButtonOptions{
		Icon:     icon,
		Label:    label,
		Size:     28,
		IconSize: 16,
		Radius:   6,
		Enabled:  enabled,
		Action:   action,
	})
}

func (s *Sidebar) drawFeeds(c *ui.Context, theme *ui.Theme, snapshot app.Snapshot) {
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
		ui.Column(c).Fill().Gap(feedRowGap).Children(func() {
			s.drawSelectionPill(c, theme, snapshot, feeds)
			for _, feed := range feeds {
				s.drawFeed(c, theme, snapshot, feed)
			}
		})
	})
}

// drawSelectionPill draws the rounded highlight behind the selected feed.
// It slides from the previously selected row on a spring, so quick h/l runs
// keep moving instead of jumping.
func (s *Sidebar) drawSelectionPill(
	c *ui.Context,
	theme *ui.Theme,
	snapshot app.Snapshot,
	feeds []app.Feed,
) {
	index := slices.IndexFunc(feeds, func(feed app.Feed) bool {
		return feed.ID == snapshot.CurrentFeedID
	})
	if index == -1 {
		return
	}
	pill := ui.Box(c).
		Key("selection-pill").
		Absolute().
		Left(0).
		Right(0).
		Height(feedRowHeight).
		Radius(6).
		Background(theme.Accent.Alpha(0.22)).
		ZIndex(-1).
		IgnoreInput(true).
		Label("选中订阅源")
	target := float32(index) * (feedRowHeight + feedRowGap)
	if snapshot.ReduceMotion {
		pill.Top(pill.Jump("y", target))
	} else {
		pill.Top(pill.AnimateTo("y", target, selectionSlide))
	}
}

func (s *Sidebar) drawFeed(
	c *ui.Context,
	theme *ui.Theme,
	snapshot app.Snapshot,
	feed app.Feed,
) {
	selected := snapshot.CurrentFeedID == feed.ID
	row := ui.ButtonBase(c).
		Key("feed-"+formatFeedID(feed.ID)).
		Grow(1).
		Height(feedRowHeight).
		Padding(10, 8).
		Radius(6).
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
				ui.Box(c).PaddingStart(8).Children(func() {
					ui.Badge(c, unreadText(feed.Unread))
				})
			}
		})

	if row.Clicked() {
		target := feed.ID
		if selected {
			target = aggregateFeedID
		}
		s.runAsync(func() { s.callbacks.Select(target) })
	}
	row.ContextMenu(func(menu *ui.Menu) {
		if menu.Item("打开").Chosen() {
			s.runAsync(func() { s.callbacks.Select(feed.ID) })
		}
		if menu.Item("编辑").Chosen() {
			s.editFeed(feed.ID)
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

func (s *Sidebar) drawFeedIcon(c *ui.Context, theme *ui.Theme, feed app.Feed) {
	icon, ok := s.cachedIcon(feed)
	switch {
	case ok && !icon.missing:
		ui.Image(c, icon.source).Size(16, 16)
	case ok && icon.missing:
		ui.Icon(c, appicon.RSS).Size(16, 16).TextColor(theme.TextMuted)
	default:
		ui.Icon(c, appicon.RSS).Size(16, 16).TextColor(theme.TextMuted)
		s.loadIcon(feed.ID)
	}
}

func (s *Sidebar) loadIcon(feedID int64) {
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

func (s *Sidebar) cachedIcon(feed app.Feed) (sidebarIcon, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	icon, ok := s.icons[feed.ID]
	if !ok || icon.feedURL != feed.IconURL {
		return sidebarIcon{}, false
	}
	if icon.missing {
		return icon, true
	}
	return icon, time.Since(icon.fetched) < iconTTL
}

func (s *Sidebar) feedByID(feedID int64) (app.Feed, bool) {
	for _, feed := range s.model.Snapshot().Feeds {
		if feed.ID == feedID {
			return feed, true
		}
	}
	return app.Feed{}, false
}

func (s *Sidebar) removeFeed(feedID int64) {
	s.mu.Lock()
	window := s.window
	callbacks := s.callbacks
	s.mu.Unlock()
	s.runAsync(func() { callbacks.Remove(window, feedID) })
}

// editFeed asks the application to open the subscription editor for feedID.
func (s *Sidebar) editFeed(feedID int64) {
	s.mu.Lock()
	window := s.window
	callbacks := s.callbacks
	s.mu.Unlock()
	s.runAsync(func() { callbacks.Edit(window, feedID) })
}

func (s *Sidebar) runAsync(operation func()) {
	if operation == nil {
		return
	}
	go operation()
}

func (s *Sidebar) beginShow(hovered bool) {
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

func (s *Sidebar) overlayState() sidebarOverlay {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.overlay
}

func (s *Sidebar) setOverlayState(state sidebarOverlay) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overlay = state
}

func (s *Sidebar) setEntered(entered bool) {
	s.mu.Lock()
	s.entered = entered
	s.mu.Unlock()
}

func (s *Sidebar) scheduleHide() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hideTimer != nil {
		s.hideTimer.Stop()
	}
	s.hideTimer = time.AfterFunc(hideDelay, func() {
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

func visibleSidebarFeeds(feeds []app.Feed, filter string) []app.Feed {
	needle := strings.ToLower(strings.TrimSpace(filter))
	if needle == "" {
		return feeds
	}
	return slices.Clone(slices.DeleteFunc(feeds, func(feed app.Feed) bool {
		return !strings.Contains(strings.ToLower(feed.Title), needle) &&
			!strings.Contains(strings.ToLower(feed.URL), needle)
	}))
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
