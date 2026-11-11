package gui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/service"
	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

const (
	postListPaneSaveDelay = 500 * time.Millisecond
	postListTextSize      = 13
	postListSearchDelay   = 300 * time.Millisecond
	postListPageSize      = 100
	postListTimeRefresh   = 30 * time.Second
)

// PostListCallbacks connects the native article list to application actions.
type PostListCallbacks struct {
	List           func(service.PostQuery) (service.PostPage, error)
	Select         func(post store.PostSummary)
	ToggleStar     func(feedID int64, guid string) (store.PostSummary, error)
	MarkAllRead    func(feedID int64)
	SetPanePercent func(percent float32)
	Notify         func(level, message string)
}

// PostList renders articles with the WebView list's visual contract.
type PostList struct {
	mu         sync.Mutex
	model      *appstatus.Model
	callbacks  PostListCallbacks
	invalidate func()
	window     *mygo.Window

	panePercent   float32
	paneSaveTimer *time.Timer

	posts         []store.PostSummary
	total         int
	loading       bool
	loadedVersion uint64
	loadedFeed    int64
	loadedQuery   service.PostQuery
	loadVersion   int
	forceReload   bool

	appliedPostVersion uint64
	query              service.PostQuery
	search             string
	searchOpen         bool
	searchTimer        *time.Timer
	pendingScroll      string
}

// NewPostList creates the native article list.
func NewPostList(model *appstatus.Model, callbacks PostListCallbacks) *PostList {
	snapshot := model.Snapshot()
	query := service.PostQuery{
		FeedID:    snapshot.CurrentFeedID,
		Limit:     postListPageSize,
		SortOrder: service.SortOrderDescending,
	}
	return &PostList{
		model:       model,
		callbacks:   callbacks,
		panePercent: snapshot.ListPercent,
		query:       query,
	}
}

// Attach connects the list to the main window.
func (l *PostList) Attach(window *mygo.Window) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.window = window
}

// SetInvalidate installs the callback used after asynchronous updates.
func (l *PostList) SetInvalidate(invalidate func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.invalidate = invalidate
}

// PostListTheme returns the theme used by the article list and splitter.
func (l *PostList) PostListTheme(c *ui.Context) *ui.Theme {
	snapshot := l.model.Snapshot()
	return postListTheme(c, snapshot.Theme, snapshot.FontSize)
}

// SyncLayout copies the mirrored list layout into its splitter.
func (l *PostList) SyncLayout() {
	snapshot := l.model.Snapshot()
	l.mu.Lock()
	l.panePercent = snapshot.ListPercent
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) splitPercent() *float32 {
	return &l.panePercent
}

func (l *PostList) savePanePercent() {
	l.mu.Lock()
	percent := l.panePercent
	if l.paneSaveTimer != nil {
		l.paneSaveTimer.Stop()
	}
	callbacks := l.callbacks
	invalidate := l.invalidate
	l.paneSaveTimer = time.AfterFunc(postListPaneSaveDelay, func() {
		l.mu.Lock()
		l.paneSaveTimer = nil
		l.mu.Unlock()
		callbacks.SetPanePercent(percent)
	})
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

// View builds the article list content.
func (l *PostList) View(c *ui.Context) {
	snapshot := l.model.Snapshot()
	theme := postListTheme(c, snapshot.Theme, snapshot.FontSize)
	l.syncState(snapshot)
	l.installShortcuts(c)

	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Background(postListPalette(*theme).Background).Children(func() {
			l.drawHeader(c, theme, snapshot)
			l.drawPosts(c, theme, snapshot)
		})
	})
}

func (l *PostList) drawHeader(c *ui.Context, theme *ui.Theme, snapshot appstatus.Snapshot) {
	title := "聚合源"
	for _, feed := range snapshot.Feeds {
		if feed.ID == snapshot.CurrentFeedID {
			title = feed.Title
			break
		}
	}
	palette := postListPalette(*theme)

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
				l.badge(c, theme, strconv.Itoa(postListUnread(snapshot))+" 未读", palette.Secondary, palette.SecondaryForeground)
			})

			searchLabel := "搜索标题"
			if l.searchOpen {
				searchLabel = "关闭标题搜索"
			}
			l.iconButton(c, theme, postListSearch, searchLabel, l.searchOpen, func() {
				l.setSearchOpen(!l.searchOpen)
			})
			sortIcon := postListSortDescending
			sortLabel := "按时间升序排列"
			if l.query.SortOrder == service.SortOrderAscending {
				sortIcon = postListSortAscending
				sortLabel = "按时间降序排列"
			}
			l.iconButton(c, theme, sortIcon, sortLabel, false, func() {
				l.setSortOrder()
			})
			ui.Row(c).AlignItems(ui.Center).Gap(l.scaled(theme, 4)).Children(func() {
				l.iconButton(c, theme, postHeaderMail, "只看未读", l.query.UnreadOnly, func() {
					l.setUnreadOnly(!l.query.UnreadOnly)
				})
				l.iconButton(c, theme, postListStar, "只看星标", l.query.StarredOnly, func() {
					l.setStarredOnly(!l.query.StarredOnly)
				})
			})
			l.iconButton(c, theme, postListCheckAll, "全部标记为已读", false, func() {
				l.markAllRead(snapshot.CurrentFeedID)
			})
		})

		if l.searchOpen {
			ui.TextInput(c, &l.search).
				Margin(l.scaled(theme, 6), 0, 0, 0).
				Height(l.scaled(theme, 28)).
				Radius(l.scaled(theme, 6)).
				Padding(l.scaled(theme, 2), l.scaled(theme, 8)).
				FontSize(l.scaled(theme, 12)).
				FixedLineHeight(l.scaled(theme, 16)).
				Label("搜索文章标题").
				Placeholder("搜索标题").
				PlaceholderColor(palette.TextMuted).
				AutoFocus().
				OnChange(func() { l.setSearch(l.search) })
		}
	})
	ui.Box(c).FillWidth().Height(1).Background(palette.Border)
}

func (l *PostList) drawPosts(c *ui.Context, theme *ui.Theme, snapshot appstatus.Snapshot) {
	ui.Scroll(c).Grow(1).Padding(l.scaled(theme, 8), l.scaled(theme, 8)).Children(func() {
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
				l.drawPost(c, theme, snapshot, post)
			}
			if len(l.posts) < l.total {
				l.drawLoadMore(c, theme)
			}
		})
	})
}

func (l *PostList) drawPost(c *ui.Context, theme *ui.Theme, snapshot appstatus.Snapshot, post store.PostSummary) {
	palette := postListPalette(*theme)
	rowKey := postRowKey(post)
	selected := snapshot.CurrentPost.FeedID == post.FeedID && snapshot.CurrentPost.GUID == post.GUID

	row := ui.ButtonBase(c).Key(rowKey).
		FillWidth().
		Padding(l.scaled(theme, 8), l.scaled(theme, 10)).
		Radius(l.scaled(theme, 6)).
		Background(postRowBackground(palette, selected)).
		Transition(ui.ElementTransition{Duration: 100 * time.Millisecond, Colors: true}).
		Label(post.Title)
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
				l.metaText(c, theme, formatRelativeTime(post.PublishedAt, snapshot.RelativeStart))
				if post.Author != "" {
					l.metaText(c, theme, post.Author).
						MaxWidth(l.scaled(theme, 128)).
						SingleLine().
						Ellipsis("…")
				}
				if post.WordCount > 0 {
					l.metaText(c, theme, formatReadingMinutes(post.WordCount)+" · "+formatWordCount(post.WordCount)).
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
			IgnoreInput(!(post.Starred || hovered))
		starHovered := star.Hovered()
		starPressed := star.Pressed()
		if starHovered || starPressed {
			star.Background(palette.Accent)
		} else {
			star.Background(ui.Transparent)
		}
		if post.Starred || hovered {
			icon := postListStar
			if post.Starred {
				icon = postListStarFilled
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
		l.runAsync(func() { l.callbacks.Select(post) })
	}
	l.scrollPending(rowKey, row)
}

func (l *PostList) metaText(c *ui.Context, theme *ui.Theme, value string) ui.Element {
	return ui.Text(c, value).
		FontSize(l.scaled(theme, 10)).
		SingleLine().
		TextColor(postListPalette(*theme).TextMuted).
		FontFeatures("tnum")
}

func (l *PostList) scrollPending(rowKey string, row ui.Element) {
	l.mu.Lock()
	pending := l.pendingScroll == rowKey
	if pending {
		l.pendingScroll = ""
	}
	l.mu.Unlock()
	if pending {
		row.ScrollIntoView(ui.ScrollIntoViewOptions{Block: ui.ScrollNearest})
	}
}

func (l *PostList) installShortcuts(c *ui.Context) {
	c.After(postListTimeRefresh)
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

// SelectNeighbour selects and reveals the next loaded article without wrapping.
func (l *PostList) SelectNeighbour(delta int, onlyUnread bool) {
	snapshot := l.model.Snapshot()
	direction := 1
	if delta < 0 {
		direction = -1
	}

	l.mu.Lock()
	if len(l.posts) == 0 {
		l.mu.Unlock()
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
		l.mu.Unlock()
		return
	}
	post := l.posts[index]
	l.pendingScroll = postRowKey(post)
	callbacks := l.callbacks
	invalidate := l.invalidate
	l.mu.Unlock()

	if callbacks.Select != nil {
		callbacks.Select(post)
	}
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) activateSelection() {
	snapshot := l.model.Snapshot()
	l.mu.Lock()
	if len(l.posts) == 0 {
		l.mu.Unlock()
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
	l.mu.Unlock()
	if callbacks.Select != nil {
		callbacks.Select(post)
	}
}

func (l *PostList) toggleSelectedStar() {
	snapshot := l.model.Snapshot()
	l.mu.Lock()
	var post store.PostSummary
	found := false
	for _, candidate := range l.posts {
		if candidate.FeedID == snapshot.CurrentPost.FeedID &&
			candidate.GUID == snapshot.CurrentPost.GUID {
			post = candidate
			found = true
			break
		}
	}
	l.mu.Unlock()
	if found {
		l.toggleStar(post)
	}
}

func (l *PostList) drawLoadMore(c *ui.Context, theme *ui.Theme) {
	palette := postListPalette(*theme)
	label := fmt.Sprintf("加载更多（还有 %d 篇）", l.total-len(l.posts))
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

func (l *PostList) drawEmpty(c *ui.Context, theme *ui.Theme) {
	palette := postListPalette(*theme)
	filtered := l.query.UnreadOnly || l.query.StarredOnly
	title := "暂时没有文章"
	hint := "刷新订阅源后再来看看。"
	if filtered {
		title = "没有符合条件的文章"
		hint = "关掉筛选，或者刷新这个订阅源。"
	}
	ui.Column(c).Fill().Center().Gap(l.scaled(theme, 8)).
		Padding(l.scaled(theme, 40), l.scaled(theme, 24)).Children(func() {
		ui.Icon(c, postListFile).
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

func (l *PostList) drawSkeletonRow(c *ui.Context, theme *ui.Theme) {
	palette := postListPalette(*theme)
	ui.Column(c).FillWidth().Gap(l.scaled(theme, 8)).
		Padding(l.scaled(theme, 8), l.scaled(theme, 10)).Children(func() {
		ui.Box(c).Height(l.scaled(theme, 14)).WidthPercent(75.0).Background(palette.Primary.Alpha(0.1))
		ui.Box(c).Height(l.scaled(theme, 10)).FillWidth().Background(palette.Primary.Alpha(0.1))
		ui.Box(c).Height(l.scaled(theme, 10)).WidthPercent(83.0).Background(palette.Primary.Alpha(0.1))
	})
}

func (l *PostList) iconButton(
	c *ui.Context,
	theme *ui.Theme,
	icon *ui.SVG,
	label string,
	active bool,
	action func(),
) ui.Element {
	palette := postListPalette(*theme)
	var button ui.Element
	button = ui.ButtonBase(c).
		Size(l.scaled(theme, 28), l.scaled(theme, 28)).
		Radius(l.scaled(theme, 6)).Cursor(ui.CursorPointer).
		Label(label).Tooltip(label)
	hovered := button.Hovered()
	pressed := button.Pressed()
	iconColor := palette.Foreground
	if active {
		iconColor = palette.Primary
	}
	if hovered || pressed {
		iconColor = palette.AccentForeground
	}
	button.Children(func() {
		ui.Icon(c, icon).
			Size(l.scaled(theme, 14), l.scaled(theme, 14)).
			TextColor(iconColor)
	})
	if pressed || hovered {
		button.Background(palette.Accent)
	} else {
		button.Background(ui.Transparent)
	}
	if button.Clicked() {
		action()
	}
	return button
}

func (l *PostList) badge(c *ui.Context, theme *ui.Theme, label string, background, foreground ui.Color) {
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

func (l *PostList) syncState(snapshot appstatus.Snapshot) {
	l.mu.Lock()
	if snapshot.PostVersion != l.appliedPostVersion {
		l.appliedPostVersion = snapshot.PostVersion
		for index, post := range l.posts {
			if post.FeedID == snapshot.Post.FeedID && post.GUID == snapshot.Post.GUID {
				l.posts[index].Read = snapshot.Post.Read
				l.posts[index].Starred = snapshot.Post.Starred
				break
			}
		}
	}

	changedFeed := snapshot.CurrentFeedID != l.loadedFeed
	changedQuery := l.query != l.loadedQuery
	changedPosts := snapshot.PostsVersion != l.loadedVersion
	needsLoad := l.forceReload || l.loadVersion == 0 || changedFeed || changedQuery || changedPosts
	if !needsLoad || l.loading {
		l.mu.Unlock()
		return
	}
	l.forceReload = false
	l.loadedFeed = snapshot.CurrentFeedID
	l.loadedVersion = snapshot.PostsVersion
	l.query.FeedID = snapshot.CurrentFeedID
	query := l.query
	query.FeedID = snapshot.CurrentFeedID
	query.Offset = 0
	query.Limit = postListPageSize
	l.loadedQuery = query
	l.loadVersion++
	generation := l.loadVersion
	l.loading = true
	callbacks := l.callbacks
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}

	go func() {
		page, err := callbacks.List(query)
		l.mu.Lock()
		if generation == l.loadVersion {
			if err == nil {
				l.posts = page.Posts
				l.total = page.Total
			}
			l.loading = false
		}
		invalidate := l.invalidate
		l.mu.Unlock()
		if err != nil && callbacks.Notify != nil {
			callbacks.Notify("error", "无法读取文章："+err.Error())
		}
		if invalidate != nil {
			invalidate()
		}
	}()
}

func (l *PostList) setSortOrder() {
	l.mu.Lock()
	if l.query.SortOrder == service.SortOrderAscending {
		l.query.SortOrder = service.SortOrderDescending
	} else {
		l.query.SortOrder = service.SortOrderAscending
	}
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) setUnreadOnly(active bool) {
	l.mu.Lock()
	l.query.UnreadOnly = active
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) setStarredOnly(active bool) {
	l.mu.Lock()
	l.query.StarredOnly = active
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) setSearchOpen(open bool) {
	l.mu.Lock()
	l.searchOpen = open
	if !open {
		l.search = ""
		l.query.Search = ""
		if l.searchTimer != nil {
			l.searchTimer.Stop()
			l.searchTimer = nil
		}
	}
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) setSearch(value string) {
	l.mu.Lock()
	if l.searchTimer != nil {
		l.searchTimer.Stop()
	}
	query := strings.TrimSpace(value)
	if query == "" {
		l.search = query
		l.query.Search = query
		invalidate := l.invalidate
		l.mu.Unlock()
		if invalidate != nil {
			invalidate()
		}
		return
	}
	invalidate := l.invalidate
	l.searchTimer = time.AfterFunc(postListSearchDelay, func() {
		l.mu.Lock()
		l.search = query
		l.query.Search = query
		l.searchTimer = nil
		current := l.invalidate
		l.mu.Unlock()
		if current != nil {
			current()
		}
	})
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) loadMore() {
	l.mu.Lock()
	if l.loading {
		l.mu.Unlock()
		return
	}
	query := l.query
	query.Offset = len(l.posts)
	query.Limit = postListPageSize
	l.loadVersion++
	generation := l.loadVersion
	l.loading = true
	callbacks := l.callbacks
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}

	go func() {
		page, err := callbacks.List(query)
		l.mu.Lock()
		if generation == l.loadVersion {
			if err == nil {
				l.posts = append(l.posts, page.Posts...)
				l.total = page.Total
			}
			l.loading = false
		}
		invalidate := l.invalidate
		l.mu.Unlock()
		if err != nil && callbacks.Notify != nil {
			callbacks.Notify("error", "无法读取文章："+err.Error())
		}
		if invalidate != nil {
			invalidate()
		}
	}()
}

func (l *PostList) toggleStar(post store.PostSummary) {
	l.runAsync(func() {
		if _, err := l.callbacks.ToggleStar(post.FeedID, post.GUID); err != nil && l.callbacks.Notify != nil {
			l.callbacks.Notify("error", "无法修改星标："+err.Error())
		}
	})
}

func (l *PostList) markAllRead(feedID int64) {
	l.runAsync(func() {
		if l.callbacks.MarkAllRead != nil {
			l.callbacks.MarkAllRead(feedID)
		}
		l.reload()
	})
}

func (l *PostList) reload() {
	l.mu.Lock()
	l.forceReload = true
	l.loadedVersion = 0
	invalidate := l.invalidate
	l.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

func (l *PostList) runAsync(operation func()) {
	if operation == nil {
		return
	}
	go operation()
}

func (l *PostList) scaled(theme *ui.Theme, size float32) float32 {
	return size * theme.FontSize / postListTextSize
}

func postListTheme(c *ui.Context, appearance, fontSize string) *ui.Theme {
	theme := nativeTheme(c, appearance, fontSize)
	if nativeTextSize > 0 {
		theme.FontSize *= postListTextSize / nativeTextSize
	}
	palette := postListPalette(*theme)
	theme.Background = palette.Background
	theme.Surface = ui.Transparent
	theme.SurfaceHover = palette.Accent
	theme.SurfacePressed = palette.Accent
	theme.Text = palette.Foreground
	theme.TextMuted = palette.TextMuted
	theme.Border = palette.Border
	theme.Scrollbar = palette.TextMuted.Alpha(0.4)
	return theme
}

func postListPalette(base ui.Theme) postPalette {
	if base.Dark {
		return postPalette{
			Background:          ui.Oklch(0.17, 0.005, 285),
			Foreground:          ui.Oklch(0.93, 0.002, 285),
			TextMuted:           ui.Oklch(0.68, 0.01, 285),
			Secondary:           ui.Oklch(0.25, 0.006, 285),
			Muted:               ui.Oklch(0.24, 0.006, 285),
			SecondaryForeground: ui.Oklch(0.93, 0.002, 285),
			Accent:              ui.Oklch(0.27, 0.008, 285),
			AccentForeground:    ui.Oklch(0.95, 0.002, 285),
			Border:              ui.Oklch(0.30, 0.006, 285),
			Primary:             base.Accent,
			Warning:             ui.Oklch(0.80, 0.14, 80),
		}
	}
	if base.Background == ui.Oklch(0.86, 0.025, 80) {
		return postPalette{
			Background:          ui.Oklch(0.86, 0.025, 80),
			Foreground:          ui.Oklch(0.24, 0.02, 70),
			TextMuted:           ui.Oklch(0.44, 0.022, 70),
			Secondary:           ui.Oklch(0.82, 0.025, 80),
			Muted:               ui.Oklch(0.82, 0.025, 80),
			SecondaryForeground: ui.Oklch(0.28, 0.02, 70),
			Accent:              ui.Oklch(0.80, 0.03, 78),
			AccentForeground:    ui.Oklch(0.26, 0.02, 70),
			Border:              ui.Oklch(0.77, 0.02, 78),
			Primary:             base.Accent,
			Warning:             ui.Oklch(0.66, 0.13, 65),
		}
	}
	return postPalette{
		Background:          ui.Oklch(0.995, 0, 285),
		Foreground:          ui.Oklch(0.22, 0.006, 285),
		TextMuted:           ui.Oklch(0.53, 0.01, 285),
		Secondary:           ui.Oklch(0.96, 0.002, 285),
		Muted:               ui.Oklch(0.96, 0.002, 285),
		SecondaryForeground: ui.Oklch(0.28, 0.006, 285),
		Accent:              ui.Oklch(0.94, 0.004, 285),
		AccentForeground:    ui.Oklch(0.25, 0.006, 285),
		Border:              ui.Oklch(0.90, 0.003, 285),
		Primary:             base.Accent,
		Warning:             ui.Oklch(0.78, 0.15, 78),
	}
}

type postPalette struct {
	Background          ui.Color
	Foreground          ui.Color
	TextMuted           ui.Color
	Secondary           ui.Color
	Muted               ui.Color
	SecondaryForeground ui.Color
	Accent              ui.Color
	AccentForeground    ui.Color
	Border              ui.Color
	Primary             ui.Color
	Warning             ui.Color
}

func postRowBackground(palette postPalette, selected bool) ui.Color {
	if selected {
		return palette.Accent
	}
	return palette.Background
}

func postRowKey(post store.PostSummary) string {
	return "post-" + formatPostID(post.FeedID) + "-" + post.GUID
}

func postListUnread(snapshot appstatus.Snapshot) int {
	if snapshot.CurrentFeedID == service.AggregateFeedID {
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
