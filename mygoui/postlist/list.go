// Package postlist renders the native article list and its paging controls.
package postlist

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
	appformat "github.com/sdttttt/ressu/mygoui/format"
	appicon "github.com/sdttttt/ressu/mygoui/icon"
	guitheme "github.com/sdttttt/ressu/mygoui/theme"
	"github.com/sdttttt/ressu/mygoui/widget"
)

const (
	paneSaveDelay  = 500 * time.Millisecond
	textSize       = 13
	searchDelay    = 300 * time.Millisecond
	pageSize       = 20
	timeRefresh    = 30 * time.Second
	reloadDebounce = 300 * time.Millisecond
)

// Callbacks connects the native article list to application actions.
type Callbacks struct {
	List           func(app.PostQuery) (app.PostPage, error)
	Select         func(post app.PostSummary, direction int)
	ToggleRead     func(feedID int64, guid string) (app.PostSummary, error)
	ToggleStar     func(feedID int64, guid string) (app.PostSummary, error)
	MarkAllRead    func(parent *mygo.Window, feedID int64)
	OpenLink       func(link string) error
	CopyLink       func(link string)
	SetPanePercent func(percent float32)
	Notify         func(level, message string)
}

// List renders articles with the WebView list's visual contract.
//
// All state below is owned by the main UI thread: View, shortcuts, and draw
// callbacks run there, and asynchronous completions arrive through submit,
// which marshals them onto the main thread. mu guards only the wiring fields
// set once at startup, which events from service goroutines may read before
// Attach runs.
type List struct {
	mu         sync.Mutex
	model      *app.Model
	callbacks  Callbacks
	invalidate func()
	window     *mygo.Window
	pending    []func()

	panePercent   float32
	paneSaveTimer *time.Timer

	posts         []app.PostSummary
	total         int
	loading       bool
	loadedVersion uint64
	loadedFeed    int64
	loadedQuery   app.PostQuery
	loadVersion   int
	forceReload   bool

	appliedPostVersion uint64
	query              app.PostQuery
	search             string
	searchOpen         bool
	searchTimer        *time.Timer
	searchScope        int
	reloadTimer        *time.Timer
	pendingScroll      string
	scroll             ui.ScrollState

	anchorKey         string
	preserveScroll    bool
	pendingRestoreKey string
	pendingRestoreY   float32
	reloadLimit       int
}

// New creates the native article list.
func New(model *app.Model, callbacks Callbacks) *List {
	snapshot := model.Snapshot()
	query := app.PostQuery{
		FeedID:     snapshot.CurrentFeedID,
		UnreadOnly: true,
		Limit:      pageSize,
		SortOrder:  app.SortOrderDescending,
		SortBy:     app.SortField(snapshot.PostSort),
	}
	return &List{
		model:       model,
		callbacks:   callbacks,
		panePercent: snapshot.ListPercent,
		query:       query,
	}
}

// Attach connects the list to the main window.
func (l *List) Attach(window *mygo.Window) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.window = window
}

// SetInvalidate installs the callback used after asynchronous updates.
func (l *List) SetInvalidate(invalidate func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.invalidate = invalidate
}

// Theme returns the theme used by the article list and splitter.
func (l *List) Theme(c *ui.Context) *ui.Theme {
	snapshot := l.model.Snapshot()
	return guitheme.PostTheme(c, snapshot.Theme, snapshot.FontSize)
}

// SyncLayout copies the mirrored list layout into its splitter.
func (l *List) SyncLayout() {
	snapshot := l.model.Snapshot()
	l.submit(func() {
		l.panePercent = snapshot.ListPercent
		l.redraw()
	})
}

func (l *List) SplitPercent() *float32 {
	return &l.panePercent
}

// SavePanePercent stores the splitter position once the drag settles.
func (l *List) SavePanePercent() {
	percent := l.panePercent
	if l.paneSaveTimer != nil {
		l.paneSaveTimer.Stop()
	}
	callbacks := l.callbacks
	l.paneSaveTimer = time.AfterFunc(paneSaveDelay, func() {
		l.submit(func() {
			l.paneSaveTimer = nil
			callbacks.SetPanePercent(percent)
		})
	})
	l.redraw()
}

// View builds the article list content.
func (l *List) View(c *ui.Context) {
	l.flushPending()
	snapshot := l.model.Snapshot()
	theme := guitheme.PostTheme(c, snapshot.Theme, snapshot.FontSize)
	l.syncState(snapshot)
	l.installShortcuts(c)

	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Background(guitheme.NewPostPalette(*theme).Background).Children(func() {
			l.drawHeader(c, theme, snapshot)
			l.drawPosts(c, theme, snapshot)
		})
	})
}

func (l *List) drawHeader(c *ui.Context, theme *ui.Theme, snapshot app.Snapshot) {
	title := "聚合源"
	for _, feed := range snapshot.Feeds {
		if feed.ID == snapshot.CurrentFeedID {
			title = feed.Title
			break
		}
	}
	palette := guitheme.NewPostPalette(*theme)

	ui.Column(c).FillWidth().Padding(l.scaled(theme, 8), l.scaled(theme, 12)).Children(func() {
		ui.Text(c, title).
			FontSize(l.scaled(theme, 14)).
			FixedLineHeight(l.scaled(theme, 20)).
			FontWeight(600).
			SingleLine().
			Ellipsis("…")
		ui.Row(c).Margin(l.scaled(theme, 4), 0, 0, 0).Wrap().
			GapX(l.scaled(theme, 4)).GapY(l.scaled(theme, 4)).
			AlignItems(ui.Center).Children(func() {
			ui.Row(c).Grow(1).AlignItems(ui.Center).Gap(l.scaled(theme, 6)).Children(func() {
				ui.Text(c, strconv.Itoa(l.total)+" 篇").
					FontSize(l.scaled(theme, 11)).
					FixedLineHeight(l.scaled(theme, 15)).
					FontFeatures("tnum").
					TextColor(palette.TextMuted)
				ui.Box(c).Width(1).Height(l.scaled(theme, 12)).Background(palette.Border)
				l.badge(c, theme, strconv.Itoa(currentUnread(snapshot))+" 未读", palette.Secondary, palette.SecondaryForeground)
			})

			searchLabel := "搜索文章"
			if l.searchOpen {
				searchLabel = "关闭搜索"
			}
			l.iconButton(c, theme, appicon.Search, searchLabel, l.searchOpen, func() {
				l.setSearchOpen(!l.searchOpen)
			})
			sortIcon := appicon.SortDescending
			sortLabel := "按时间升序排列"
			if l.query.SortOrder == app.SortOrderAscending {
				sortIcon = appicon.SortAscending
				sortLabel = "按时间降序排列"
			}
			l.iconButton(c, theme, sortIcon, sortLabel, false, func() {
				l.setSortOrder()
			})
			ui.Row(c).AlignItems(ui.Center).Gap(l.scaled(theme, 4)).Children(func() {
				unreadLabel := "只看未读"
				if !l.query.UnreadOnly {
					unreadLabel = "显示全部"
				}
				l.iconButton(c, theme, appicon.Mail, unreadLabel, !l.query.UnreadOnly, func() {
					l.setUnreadOnly(!l.query.UnreadOnly)
				})
				l.iconButton(c, theme, appicon.Star, "只看星标", l.query.StarredOnly, func() {
					l.setStarredOnly(!l.query.StarredOnly)
				})
			})
			l.iconButton(c, theme, appicon.CheckAll, "全部标记为已读", false, func() {
				l.markAllRead(snapshot.CurrentFeedID)
			})
		})

		if l.searchOpen {
			// The post theme makes the control surfaces transparent for
			// the list; the segmented control needs the standard surfaces
			// to read as a control, as in the settings window.
			base := guitheme.New(c, snapshot.Theme, snapshot.FontSize)
			base.FontSize = theme.FontSize
			ui.Themed(c, base).Children(func() {
				scope := ui.Segmented(c, &l.searchScope, "标题", "全文").
					Label("搜索范围").
					Margin(l.scaled(theme, 6), 0, 0, 0)
				if scope.Changed() {
					l.setSearchScope(l.searchScope)
				}
			})
			inputLabel := "搜索文章标题"
			placeholder := "搜索标题"
			if l.searchScope == 1 {
				inputLabel = "搜索文章全文"
				placeholder = "搜索全文"
			}
			ui.TextInput(c, &l.search).
				Margin(l.scaled(theme, 6), 0, 0, 0).
				Height(l.scaled(theme, 28)).
				Radius(l.scaled(theme, 6)).
				Padding(l.scaled(theme, 2), l.scaled(theme, 8)).
				FontSize(l.scaled(theme, 12)).
				FixedLineHeight(l.scaled(theme, 16)).
				Label(inputLabel).
				Placeholder(placeholder).
				PlaceholderColor(palette.TextMuted).
				AutoFocus().
				OnChange(func() { l.setSearch(l.search) })
		}
	})
	ui.Box(c).FillWidth().Height(1).Background(palette.Border)
}

func (l *List) drawPosts(c *ui.Context, theme *ui.Theme, snapshot app.Snapshot) {
	scrollEl := ui.Scroll(c).Grow(1).TrackScroll(&l.scroll).Padding(l.scaled(theme, 8), l.scaled(theme, 8))
	viewport := scrollEl.Bounds()
	anchorKey := ""
	scrollEl.Children(func() {
		if l.loading && len(l.posts) == 0 {
			ui.Column(c).FillWidth().Gap(l.scaled(theme, 8)).Children(func() {
				for range 5 {
					l.drawSkeletonRow(c, theme)
				}
			})
			return
		}

		if len(l.posts) == 0 {
			l.drawEmpty(c, theme)
			return
		}

		ui.Column(c).FillWidth().Gap(l.scaled(theme, 4)).Children(func() {
			for _, post := range l.posts {
				l.drawPost(c, theme, snapshot, post, viewport, &anchorKey)
			}
			if l.matchCount() < l.total {
				l.drawLoadMore(c, theme)
			}
		})
	})
	if anchorKey != "" {
		l.anchorKey = anchorKey
	}
	l.autoLoadMore(theme)
}

// autoLoadMore starts the next page once the user scrolls close to the
// bottom of the loaded list, so paging no longer depends on clicking the
// load-more button. The button stays as the fallback for viewports the
// content does not overflow, where no scroll event can reach the bottom.
func (l *List) autoLoadMore(theme *ui.Theme) {
	loading := l.loading
	hasMore := l.matchCount() < l.total
	threshold := l.scaled(theme, 72)
	nearBottom := l.scroll.MaxY > 0 && l.scroll.Y >= l.scroll.MaxY-threshold
	if !loading && hasMore && nearBottom {
		l.loadMore()
	}
}

func (l *List) drawPost(c *ui.Context, theme *ui.Theme, snapshot app.Snapshot, post app.PostSummary, viewport ui.Rect, frameAnchor *string) {
	palette := guitheme.NewPostPalette(*theme)
	rowKey := postRowKey(post)
	selected := snapshot.CurrentPost.FeedID == post.FeedID && snapshot.CurrentPost.GUID == post.GUID

	row := ui.ButtonBase(c).Key(rowKey).
		FillWidth().
		Padding(l.scaled(theme, 8), l.scaled(theme, 10)).
		Radius(l.scaled(theme, 6)).
		Background(postRowBackground(palette, selected)).
		Transition(ui.ElementTransition{Duration: 100 * time.Millisecond, Colors: true}).
		Label(post.Title)
	if *frameAnchor == "" {
		bounds := row.Bounds()
		if viewport.H > 0 && bounds.H > 0 && bounds.Y+bounds.H > viewport.Y {
			*frameAnchor = rowKey
		}
	}
	hovered := row.Hovered()
	if hovered {
		row.Background(palette.Accent.Alpha(0.6))
	} else {
		row.Background(postRowBackground(palette, selected))
	}
	if row.Pressed() {
		row.Scale(0.98, 0.98)
	}
	row.Children(func() {
		ui.Column(c).FillWidth().Children(func() {
			title := ui.Text(c, post.Title).
				FontSize(l.scaled(theme, 13)).
				LineHeight(1.375).
				MaxLines(2).
				Ellipsis("…").
				Wrap()
			if post.Read {
				title.TextColor(palette.TextMuted)
			} else {
				title.FontWeight(600)
			}

			if post.Summary != "" {
				ui.Text(c, post.Summary).
					Margin(l.scaled(theme, 4), 0, 0, 0).
					FontSize(l.scaled(theme, 11)).
					LineHeight(1.625).
					MaxLines(3).
					Ellipsis("…").
					Wrap().
					TextColor(palette.TextMuted)
			}

			ui.Row(c).Margin(l.scaled(theme, 6), 0, 0, 0).AlignItems(ui.Center).
				Gap(l.scaled(theme, 8)).
				Padding(0, l.scaled(theme, 28), 0, 0).Children(func() {
				if !post.Read {
					ui.Box(c).Shrink(0).
						Size(l.scaled(theme, 6), l.scaled(theme, 6)).
						Radius(l.scaled(theme, 3)).
						Background(palette.Primary)
				}
				l.metaText(c, theme, appformat.RelativeTime(post.PublishedAt, snapshot.RelativeStart))
				if post.Author != "" {
					l.metaText(c, theme, post.Author).
						MaxWidth(l.scaled(theme, 128)).
						SingleLine().
						Ellipsis("…")
				}
				if post.WordCount > 0 {
					l.metaText(c, theme, appformat.ReadingMinutes(post.WordCount)+" · "+appformat.WordCount(post.WordCount)).
						Shrink(1).
						Ellipsis("…")
				}
			})
		})

		star := ui.ButtonBase(c).Absolute().
			Bottom(l.scaled(theme, 4)).Right(l.scaled(theme, 4)).
			Size(l.scaled(theme, 24), l.scaled(theme, 24)).
			Radius(l.scaled(theme, 6)).Cursor(ui.CursorPointer).
			Label(map[bool]string{true: "取消星标", false: "加星标"}[post.Starred]).
			IgnoreInput(!post.Starred && !hovered)
		starHovered := star.Hovered()
		starPressed := star.Pressed()
		if starHovered || starPressed {
			star.Background(palette.Accent)
		} else {
			star.Background(ui.Transparent)
		}
		if post.Starred || hovered {
			icon := appicon.Star
			if post.Starred {
				icon = appicon.StarFilled
			}
			iconColor := palette.Foreground
			if post.Starred {
				iconColor = palette.Warning
			}
			if starHovered || starPressed {
				iconColor = palette.AccentForeground
			}
			star.Children(func() {
				ui.Icon(c, icon).
					Size(l.scaled(theme, 14), l.scaled(theme, 14)).
					TextColor(iconColor)
			})
		} else {
			star.Opacity(0)
		}
		if star.Clicked() {
			l.toggleStar(post)
		}
	})
	if row.Clicked() {
		selected := post
		l.runAsync(func() {
			snapshot := l.model.Snapshot()
			l.callbacks.Select(selected, l.directionFor(snapshot.CurrentPost, selected))
		})
	}
	row.ContextMenu(func(menu *ui.Menu) {
		readLabel := "标记为已读"
		if post.Read {
			readLabel = "标记为未读"
		}
		if menu.Item(readLabel).Chosen() {
			l.toggleRead(post)
		}
		starLabel := "加星标"
		if post.Starred {
			starLabel = "取消星标"
		}
		if menu.Item(starLabel).Chosen() {
			l.toggleStar(post)
		}
		menu.Separator()
		if menu.Item("复制链接").Disabled(post.URL == "").Chosen() {
			l.copyLink(post)
		}
		if menu.Item("浏览器打开").Disabled(post.URL == "").Chosen() {
			l.openOriginal(post)
		}
	})
	l.scrollPending(rowKey, row)
}

func (l *List) metaText(c *ui.Context, theme *ui.Theme, value string) ui.Element {
	return ui.Text(c, value).
		FontSize(l.scaled(theme, 10)).
		SingleLine().
		TextColor(guitheme.NewPostPalette(*theme).TextMuted).
		FontFeatures("tnum")
}

func (l *List) scrollPending(rowKey string, row ui.Element) {
	pending := l.pendingScroll == rowKey
	if pending {
		l.pendingScroll = ""
	}
	if pending {
		row.ScrollIntoView(ui.ScrollIntoViewOptions{Block: ui.ScrollNearest})
	}
}

func (l *List) installShortcuts(c *ui.Context) {
	c.After(timeRefresh)
	if c.FocusTakesText() {
		return
	}
	c.OnWindowShortcut(0, ui.KeyJ, func() { l.SelectNeighbour(1, false) })
	c.OnWindowShortcut(0, ui.KeyK, func() { l.SelectNeighbour(-1, false) })
	c.OnWindowShortcut(0, ui.KeyN, func() { l.SelectNeighbour(1, true) })
	c.OnWindowShortcut(0, ui.KeyP, func() { l.SelectNeighbour(-1, true) })
	c.OnWindowShortcut(0, ui.KeyEnter, l.activateSelection)
	c.OnWindowShortcut(0, ui.KeyS, l.toggleSelectedStar)
}

// directionFor reports where post sits relative to the current selection in
// the loaded list: 1 later, -1 earlier, 0 unknown. It runs on the main
// thread.
func (l *List) directionFor(current app.PostSelection, post app.PostSummary) int {
	return selectionDirection(l.posts, current, post)
}

// selectionDirection compares the selected post against the current
// selection in the loaded list: 1 when the post sits later in the list, -1
// when earlier, and 0 when there is no current article to compare or either
// post is missing from the loaded page.
func selectionDirection(posts []app.PostSummary, current app.PostSelection, post app.PostSummary) int {
	if current.GUID == "" {
		return 0
	}
	currentIndex, postIndex := -1, -1
	for index, item := range posts {
		if item.FeedID == current.FeedID && item.GUID == current.GUID {
			currentIndex = index
		}
		if item.FeedID == post.FeedID && item.GUID == post.GUID {
			postIndex = index
		}
	}
	switch {
	case currentIndex < 0 || postIndex < 0:
		return 0
	case postIndex > currentIndex:
		return 1
	case postIndex < currentIndex:
		return -1
	default:
		return 0
	}
}

// SelectNeighbour selects and reveals the next loaded article without
// wrapping. It runs on the main thread.
func (l *List) SelectNeighbour(delta int, onlyUnread bool) {
	snapshot := l.model.Snapshot()
	direction := 1
	if delta < 0 {
		direction = -1
	}

	if len(l.posts) == 0 {
		return
	}
	current := -1
	for index, post := range l.posts {
		if post.FeedID == snapshot.CurrentPost.FeedID && post.GUID == snapshot.CurrentPost.GUID {
			current = index
			break
		}
	}
	index := 0
	if current >= 0 {
		index = current + direction
	} else if direction < 0 {
		index = len(l.posts) - 1
	}
	for index >= 0 && index < len(l.posts) {
		if !onlyUnread || !l.posts[index].Read {
			break
		}
		index += direction
	}
	if index < 0 || index >= len(l.posts) {
		return
	}
	post := l.posts[index]
	l.pendingScroll = postRowKey(post)
	callbacks := l.callbacks

	if callbacks.Select != nil {
		selectDirection := 0
		if current >= 0 {
			if index > current {
				selectDirection = 1
			} else {
				selectDirection = -1
			}
		}
		callbacks.Select(post, selectDirection)
	}
	l.redraw()
}

func (l *List) activateSelection() {
	snapshot := l.model.Snapshot()
	if len(l.posts) == 0 {
		return
	}
	post := l.posts[0]
	for _, candidate := range l.posts {
		if candidate.FeedID == snapshot.CurrentPost.FeedID &&
			candidate.GUID == snapshot.CurrentPost.GUID {
			post = candidate
			break
		}
	}
	callbacks := l.callbacks
	if callbacks.Select != nil {
		callbacks.Select(post, 0)
	}
}

func (l *List) toggleSelectedStar() {
	snapshot := l.model.Snapshot()
	var post app.PostSummary
	found := false
	for _, candidate := range l.posts {
		if candidate.FeedID == snapshot.CurrentPost.FeedID &&
			candidate.GUID == snapshot.CurrentPost.GUID {
			post = candidate
			found = true
			break
		}
	}
	if found {
		l.toggleStar(post)
	}
}

func (l *List) drawLoadMore(c *ui.Context, theme *ui.Theme) {
	palette := guitheme.NewPostPalette(*theme)
	label := fmt.Sprintf("加载更多（还有 %d 篇）", l.total-l.matchCount())
	if l.loading {
		label = "加载中…"
	}
	button := ui.ButtonBase(c).Margin(l.scaled(theme, 4), 0, 0, 0).FillWidth().
		Height(l.scaled(theme, 36)).
		Padding(0, l.scaled(theme, 12)).
		Radius(l.scaled(theme, 6)).
		Label(label)
	hovered := button.Hovered()
	if hovered && !l.loading {
		button.Background(palette.Accent)
	}
	foreground := palette.TextMuted
	if hovered && !l.loading {
		foreground = palette.AccentForeground
	}
	if l.loading {
		button.Opacity(0.5)
	}
	button.Children(func() {
		ui.Row(c).Center().Gap(l.scaled(theme, 8)).Children(func() {
			if l.loading {
				ui.Spinner(c).
					Size(l.scaled(theme, 14), l.scaled(theme, 14)).
					TextColor(foreground)
			}
			ui.Text(c, label).
				FontSize(l.scaled(theme, 12)).
				FixedLineHeight(l.scaled(theme, 16)).
				FontWeight(500).
				TextColor(foreground)
		})
	})
	button.Disabled(l.loading)
	if button.Clicked() {
		l.loadMore()
	}
}

func (l *List) drawEmpty(c *ui.Context, theme *ui.Theme) {
	palette := guitheme.NewPostPalette(*theme)
	filtered := l.query.UnreadOnly || l.query.StarredOnly
	title := "暂时没有文章"
	hint := "刷新订阅源后再来看看。"
	if filtered {
		title = "没有符合条件的文章"
		hint = "关掉筛选，或者刷新这个订阅源。"
	}
	ui.Column(c).Fill().Center().Gap(l.scaled(theme, 8)).
		Padding(l.scaled(theme, 40), l.scaled(theme, 24)).Children(func() {
		ui.Icon(c, appicon.File).
			Size(l.scaled(theme, 36), l.scaled(theme, 36)).
			TextColor(palette.TextMuted.Alpha(0.4))
		ui.Text(c, title).
			FontSize(l.scaled(theme, 14)).
			FixedLineHeight(l.scaled(theme, 20)).
			TextColor(palette.TextMuted)
		ui.Text(c, hint).
			MaxWidth(l.scaled(theme, 224)).
			FontSize(l.scaled(theme, 12)).
			LineHeight(1.625).
			TextColor(palette.TextMuted.Alpha(0.7)).
			TextAlign(ui.Center)
	})
}

func (l *List) drawSkeletonRow(c *ui.Context, theme *ui.Theme) {
	palette := guitheme.NewPostPalette(*theme)
	ui.Column(c).FillWidth().Gap(l.scaled(theme, 8)).
		Padding(l.scaled(theme, 8), l.scaled(theme, 10)).Children(func() {
		ui.Box(c).Height(l.scaled(theme, 14)).WidthPercent(75.0).Background(palette.Primary.Alpha(0.1))
		ui.Box(c).Height(l.scaled(theme, 10)).FillWidth().Background(palette.Primary.Alpha(0.1))
		ui.Box(c).Height(l.scaled(theme, 10)).WidthPercent(83.0).Background(palette.Primary.Alpha(0.1))
	})
}

func (l *List) iconButton(
	c *ui.Context,
	theme *ui.Theme,
	icon *ui.SVG,
	label string,
	active bool,
	action func(),
) {
	widget.IconButton(c, guitheme.PostIconButtonColors(guitheme.NewPostPalette(*theme)), widget.IconButtonOptions{
		Icon:     icon,
		Label:    label,
		Size:     l.scaled(theme, 28),
		IconSize: l.scaled(theme, 14),
		Radius:   l.scaled(theme, 6),
		Enabled:  true,
		Active:   active,
		Action:   action,
	})
}

func (l *List) badge(c *ui.Context, theme *ui.Theme, label string, background, foreground ui.Color) {
	ui.Row(c).
		Height(l.scaled(theme, 16)).
		Padding(0, l.scaled(theme, 6)).
		Radius(999).Background(background).Children(func() {
		ui.Text(c, label).
			FontSize(l.scaled(theme, 10)).
			FontWeight(500).
			FontFeatures("tnum").
			SingleLine().
			TextColor(foreground)
	})
}

func (l *List) syncState(snapshot app.Snapshot) {
	if snapshot.PostSort != "" {
		l.query.SortBy = app.SortField(snapshot.PostSort)
	}
	if snapshot.PostVersion != l.appliedPostVersion {
		l.appliedPostVersion = snapshot.PostVersion
		for index, post := range l.posts {
			if post.FeedID == snapshot.Post.FeedID && post.GUID == snapshot.Post.GUID {
				l.adjustTotal(post, snapshot.Post)
				l.posts[index].Read = snapshot.Post.Read
				l.posts[index].Starred = snapshot.Post.Starred
				break
			}
		}
	}

	changedFeed := snapshot.CurrentFeedID != l.loadedFeed
	changedQuery := l.query != l.loadedQuery
	changedPosts := snapshot.PostsVersion != l.loadedVersion
	needsLoad := l.forceReload || l.loadVersion == 0 || changedFeed || changedQuery
	if !needsLoad {
		// A refresh bumps PostsVersion once per completed feed; coalesce
		// that burst into one debounced reload instead of one per feed.
		if changedPosts && !l.loading {
			l.scheduleReload()
		}
		return
	}
	if l.loading {
		return
	}
	l.forceReload = false
	l.loadedFeed = snapshot.CurrentFeedID
	l.loadedVersion = snapshot.PostsVersion
	l.query.FeedID = snapshot.CurrentFeedID
	query := l.query
	query.FeedID = snapshot.CurrentFeedID
	query.Offset = 0
	query.Limit = pageSize
	if l.reloadLimit > pageSize {
		// A data-change reload fetches enough rows to cover what the user
		// had scrolled through, so the scroll anchor survives the reload.
		query.Limit = l.reloadLimit
		l.reloadLimit = 0
	}
	l.loadedQuery = l.query
	l.loadVersion++
	generation := l.loadVersion
	l.loading = true
	callbacks := l.callbacks
	l.redraw()

	go func() {
		page, err := callbacks.List(query)
		l.submit(func() {
			if generation != l.loadVersion {
				return
			}
			if err == nil {
				l.posts = page.Posts
				l.total = page.Total
				l.applyScrollRestore()
			} else {
				l.preserveScroll = false
			}
			l.loading = false
		})
		if err != nil && callbacks.Notify != nil {
			callbacks.Notify("error", "无法读取文章："+err.Error())
		}
	}()
}

func (l *List) setSortOrder() {
	if l.query.SortOrder == app.SortOrderAscending {
		l.query.SortOrder = app.SortOrderDescending
	} else {
		l.query.SortOrder = app.SortOrderAscending
	}
	l.redraw()
}

func (l *List) setUnreadOnly(active bool) {
	l.query.UnreadOnly = active
	l.redraw()
}

func (l *List) setStarredOnly(active bool) {
	l.query.StarredOnly = active
	l.redraw()
}

func (l *List) setSearchOpen(open bool) {
	l.searchOpen = open
	if !open {
		l.search = ""
		l.query.Search = ""
		if l.searchTimer != nil {
			l.searchTimer.Stop()
			l.searchTimer = nil
		}
	}
	l.redraw()
}

func (l *List) setSearchScope(scope int) {
	l.searchScope = scope
	l.query.SearchScope = app.SearchScopeTitle
	if scope == 1 {
		l.query.SearchScope = app.SearchScopeFulltext
	}
	l.redraw()
}

func (l *List) setSearch(value string) {
	if l.searchTimer != nil {
		l.searchTimer.Stop()
	}
	query := strings.TrimSpace(value)
	if query == "" {
		l.search = query
		l.query.Search = query
		l.redraw()
		return
	}
	l.searchTimer = time.AfterFunc(searchDelay, func() {
		l.submit(func() {
			l.search = query
			l.query.Search = query
			l.searchTimer = nil
		})
	})
	l.redraw()
}

func (l *List) loadMore() {
	if l.loading {
		return
	}
	query := l.query
	query.Offset = l.matchCount()
	query.Limit = pageSize
	l.loadVersion++
	generation := l.loadVersion
	l.loading = true
	callbacks := l.callbacks
	l.redraw()

	go func() {
		page, err := callbacks.List(query)
		l.submit(func() {
			if generation != l.loadVersion {
				return
			}
			if err == nil {
				l.posts = append(l.posts, page.Posts...)
				l.total = page.Total
			}
			l.loading = false
		})
		if err != nil && callbacks.Notify != nil {
			callbacks.Notify("error", "无法读取文章："+err.Error())
		}
	}()
}

// matchCount returns how many loaded posts still satisfy the active filters.
// Read or starred state can change in place while the rows stay listed, so
// len(l.posts) is not a valid page offset under an unread/starred filter.
func (l *List) matchCount() int {
	count := 0
	for _, post := range l.posts {
		if l.matches(post) {
			count++
		}
	}
	return count
}

// matches reports whether one post satisfies the active unread/starred
// filters. Search is not part of it: a listed post's title never changes.
func (l *List) matches(post app.PostSummary) bool {
	if l.query.UnreadOnly && post.Read {
		return false
	}
	if l.query.StarredOnly && !post.Starred {
		return false
	}
	return true
}

// adjustTotal reconciles the matched total when one listed post flips
// in or out of the active filters, so the load-more counter stays accurate
// between page fetches.
func (l *List) adjustTotal(before, after app.PostSummary) {
	oldMatch := l.matches(before)
	newMatch := l.matches(after)
	switch {
	case oldMatch && !newMatch && l.total > 0:
		l.total--
	case !oldMatch && newMatch:
		l.total++
	}
}

func (l *List) toggleStar(post app.PostSummary) {
	l.runAsync(func() {
		if _, err := l.callbacks.ToggleStar(post.FeedID, post.GUID); err != nil && l.callbacks.Notify != nil {
			l.callbacks.Notify("error", "无法修改星标："+err.Error())
		}
	})
}

func (l *List) toggleRead(post app.PostSummary) {
	l.runAsync(func() {
		if _, err := l.callbacks.ToggleRead(post.FeedID, post.GUID); err != nil && l.callbacks.Notify != nil {
			l.callbacks.Notify("error", "无法修改已读状态："+err.Error())
		}
	})
}

func (l *List) copyLink(post app.PostSummary) {
	if l.callbacks.CopyLink == nil || post.URL == "" {
		return
	}
	l.callbacks.CopyLink(post.URL)
}

func (l *List) openOriginal(post app.PostSummary) {
	if l.callbacks.OpenLink == nil || post.URL == "" {
		return
	}
	l.runAsync(func() {
		if err := l.callbacks.OpenLink(post.URL); err != nil && l.callbacks.Notify != nil {
			l.callbacks.Notify("error", "无法打开链接："+err.Error())
		}
	})
}

func (l *List) markAllRead(feedID int64) {
	l.mu.Lock()
	window := l.window
	callbacks := l.callbacks
	l.mu.Unlock()
	l.runAsync(func() {
		if callbacks.MarkAllRead != nil {
			callbacks.MarkAllRead(window, feedID)
		}
		l.submit(l.reload)
	})
}

func (l *List) reload() {
	l.requestReload(false)
}

// reloadAfterDataChange reloads like reload and asks the load completion to
// restore the scroll position, so a background refresh that inserts articles
// above the viewport does not move what the user is reading.
func (l *List) reloadAfterDataChange() {
	l.requestReload(true)
}

func (l *List) requestReload(preserve bool) {
	if preserve {
		l.preserveScroll = true
		l.pendingRestoreKey = l.anchorKey
		l.pendingRestoreY = l.scroll.Y
		if count := l.matchCount(); count > pageSize {
			l.reloadLimit = count
		}
	}
	l.forceReload = true
	l.loadedVersion = 0
	l.redraw()
}

// applyScrollRestore re-anchors the viewport after a data-change
// reload. When the topmost visible post survived the reload, schedule a
// minimal scroll to it; otherwise restore the pre-reload pixel offset and
// let autoLoadMore backfill the deeper pages.
func (l *List) applyScrollRestore() {
	if !l.preserveScroll {
		return
	}
	l.preserveScroll = false
	key := l.pendingRestoreKey
	if key == "" {
		return
	}
	found := false
	for _, post := range l.posts {
		if postRowKey(post) == key {
			found = true
			break
		}
	}
	if found {
		if l.pendingScroll == "" {
			l.pendingScroll = key
		}
		return
	}
	l.scroll.Y = l.pendingRestoreY
}

// scheduleReload arms the one-shot debounce that coalesces bursts of
// posts.changed events into a single reload.
func (l *List) scheduleReload() {
	if l.reloadTimer != nil {
		return
	}
	l.reloadTimer = time.AfterFunc(reloadDebounce, func() {
		l.submit(func() {
			l.reloadTimer = nil
			l.reloadAfterDataChange()
		})
	})
}

// submit runs fn on the main thread that owns the list state and redraws
// afterwards. With a window it marshals through window.Update; windowless
// tests queue the update for the next View instead, so state still changes
// only on the thread that builds frames. A destroyed window drops the update:
// the list is gone with it.
func (l *List) submit(fn func()) {
	l.mu.Lock()
	window := l.window
	if window == nil {
		l.pending = append(l.pending, fn)
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	if window.IsDestroyed() {
		return
	}
	window.Update(fn)
}

// flushPending applies queued windowless updates at the start of a frame.
func (l *List) flushPending() {
	l.mu.Lock()
	pending := l.pending
	l.pending = nil
	l.mu.Unlock()
	for _, fn := range pending {
		fn()
	}
}

// redraw requests a frame rebuild when the window still exists.
func (l *List) redraw() {
	l.mu.Lock()
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *List) runAsync(operation func()) {
	if operation == nil {
		return
	}
	go operation()
}

func (l *List) scaled(theme *ui.Theme, size float32) float32 {
	return size * theme.FontSize / textSize
}

func postRowBackground(palette guitheme.PostPalette, selected bool) ui.Color {
	if selected {
		return palette.Accent
	}
	return palette.Background
}

func postRowKey(post app.PostSummary) string {
	return "post-" + formatPostID(post.FeedID) + "-" + post.GUID
}

func currentUnread(snapshot app.Snapshot) int {
	if snapshot.CurrentFeedID == app.AggregateFeedID {
		return snapshot.Unread
	}
	for _, feed := range snapshot.Feeds {
		if feed.ID == snapshot.CurrentFeedID {
			return feed.Unread
		}
	}
	return 0
}

func formatPostID(id int64) string {
	return strconv.FormatInt(id, 10)
}
