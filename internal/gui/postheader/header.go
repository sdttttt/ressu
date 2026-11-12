// Package postheader renders the selected article header.
package postheader

import (
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
	appformat "github.com/sdttttt/ressu/internal/gui/format"
	appicon "github.com/sdttttt/ressu/internal/gui/icon"
	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
	"github.com/sdttttt/ressu/internal/gui/widget"
)

const (
	textSize       = 13
	metaTextSize   = 10
	titleTextSize  = 18
	titleLineCount = 2
)

// Callbacks connects the native article header to application
// actions.
type Callbacks struct {
	ToggleStar      func(feedID int64, guid string) (store.PostSummary, error)
	MarkRead        func(feedID int64, guid string, read bool) error
	OpenLink        func(link string) error
	RequestSnapshot func(parent *mygo.Window)
	Notify          func(level, message string)
}

// Header renders metadata and actions for the selected article.
type Header struct {
	model     *appstatus.Model
	callbacks Callbacks

	mu     sync.Mutex
	window *mygo.Window
}

// New creates the native header above the reading pane.
func New(model *appstatus.Model, callbacks Callbacks) *Header {
	return &Header{model: model, callbacks: callbacks}
}

// Attach connects the header to the main window for child surfaces.
func (h *Header) Attach(window *mygo.Window) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.window = window
}

// View draws the selected article header, or nothing when no article is
// selected.
func (h *Header) View(c *ui.Context) {
	snapshot := h.model.Snapshot()
	post := snapshot.SelectedPost
	if post.GUID == "" {
		return
	}

	theme := guitheme.PostTheme(c, snapshot.Theme, snapshot.FontSize)
	palette := guitheme.NewPostPalette(*theme)
	ui.Themed(c, theme).FillWidth().Children(func() {
		ui.Column(c).FillWidth().
			Padding(h.scaled(theme, 12), h.scaled(theme, 18), h.scaled(theme, 12), h.scaled(theme, 18)).
			Background(palette.Background).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(h.scaled(theme, 12)).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					h.drawTitle(c, theme, post)
					h.drawMeta(c, theme, snapshot, post)
				})
				ui.Row(c).Shrink(0).AlignItems(ui.Center).Gap(h.scaled(theme, 2)).Children(func() {
					h.drawActions(c, theme, post)
				})
			})
		})
	})
	ui.Box(c).FillWidth().Height(1).Background(palette.Border)
}

func (h *Header) drawTitle(c *ui.Context, theme *ui.Theme, post store.PostSummary) {
	palette := guitheme.NewPostPalette(*theme)
	title := post.Title
	if title == "" {
		title = "无标题"
	}
	lineHeight := h.scaled(theme, 22)
	ui.Text(c, title).
		FillWidth().
		MinHeight(lineHeight).
		MaxHeight(lineHeight * titleLineCount).
		FontSize(h.scaled(theme, titleTextSize)).
		FixedLineHeight(lineHeight).
		FontWeight(600).
		MaxLines(titleLineCount).
		Wrap().
		Ellipsis("…").
		TextColor(palette.Foreground)
}

func (h *Header) drawMeta(
	c *ui.Context,
	theme *ui.Theme,
	snapshot appstatus.Snapshot,
	post store.PostSummary,
) {
	palette := guitheme.NewPostPalette(*theme)
	first := true
	separator := func() {
		if !first {
			ui.Text(c, "·").
				Shrink(0).
				FontSize(h.scaled(theme, metaTextSize)).
				TextColor(palette.TextMuted.Alpha(0.6))
		}
		first = false
	}
	meta := func(value string) {
		if value == "" {
			return
		}
		separator()
		ui.Text(c, value).
			Shrink(1).
			FontSize(h.scaled(theme, metaTextSize)).
			SingleLine().
			Ellipsis("…").
			TextColor(palette.TextMuted)
	}

	ui.Row(c).Padding(24, 0, 8, 0).FillWidth().Height(h.scaled(theme, 16)).AlignItems(ui.Center).
		Gap(h.scaled(theme, 6)).Children(func() {
		meta(appformat.DateTime(post.PublishedAt))
		meta(post.Author)
		if post.WordCount > 0 {
			meta(appformat.ReadingMinutes(post.WordCount) + " · " + appformat.WordCount(post.WordCount))
		}
		if feedTitle := selectedFeedTitle(snapshot, post.FeedID); feedTitle != "" {
			separator()
			ui.Row(c).Shrink(1).MaxWidth(h.scaled(theme, 160)).Height(h.scaled(theme, 16)).
				AlignItems(ui.Center).Padding(0, h.scaled(theme, 6)).
				Radius(h.scaled(theme, 8)).Background(palette.Secondary).Children(func() {
				ui.Text(c, feedTitle).
					FontSize(h.scaled(theme, 10)).
					SingleLine().
					Ellipsis("…").
					TextColor(palette.SecondaryForeground)
			})
		}
	})
}

func (h *Header) drawActions(c *ui.Context, theme *ui.Theme, post store.PostSummary) {
	starIcon := appicon.Star
	starLabel := "加星标"
	if post.Starred {
		starIcon = appicon.StarFilled
		starLabel = "取消星标"
	}
	h.iconButton(c, theme, starIcon, starLabel, post.Starred, true, func() {
		h.toggleStar(post)
	})

	readIcon := appicon.Mail
	readLabel := "标记为已读"
	read := true
	if post.Read {
		readIcon = appicon.MailOpen
		readLabel = "标记为未读"
		read = false
	}
	h.iconButton(c, theme, readIcon, readLabel, false, true, func() {
		h.markRead(post, read)
	})
	h.iconButton(c, theme, appicon.ExternalLink, "打开原文", false, post.URL != "", func() {
		h.openOriginal(post)
	})
	h.iconButton(c, theme, appicon.Camera, "生成快照", false, h.callbacks.RequestSnapshot != nil, func() {
		h.mu.Lock()
		window := h.window
		h.mu.Unlock()
		h.callbacks.RequestSnapshot(window)
	})
}

func (h *Header) iconButton(
	c *ui.Context,
	theme *ui.Theme,
	icon *ui.SVG,
	label string,
	accent bool,
	enabled bool,
	action func(),
) ui.Element {
	palette := guitheme.NewPostPalette(*theme)
	colors := guitheme.PostIconButtonColors(palette)
	colors.Rest = palette.TextMuted
	return widget.IconButton(c, colors, widget.IconButtonOptions{
		Icon:     icon,
		Label:    label,
		Size:     h.scaled(theme, 28),
		IconSize: h.scaled(theme, 14),
		Radius:   h.scaled(theme, 6),
		Enabled:  enabled,
		Accent:   accent,
		Action:   action,
	})
}

func (h *Header) toggleStar(post store.PostSummary) {
	if h.callbacks.ToggleStar == nil {
		return
	}
	go func() {
		if _, err := h.callbacks.ToggleStar(post.FeedID, post.GUID); err != nil {
			h.notify("更新星标失败")
		}
	}()
}

func (h *Header) markRead(post store.PostSummary, read bool) {
	if h.callbacks.MarkRead == nil {
		return
	}
	go func() {
		if err := h.callbacks.MarkRead(post.FeedID, post.GUID, read); err != nil {
			h.notify("更新阅读状态失败")
		}
	}()
}

func (h *Header) openOriginal(post store.PostSummary) {
	if h.callbacks.OpenLink == nil || post.URL == "" {
		return
	}
	go func() {
		if err := h.callbacks.OpenLink(post.URL); err != nil {
			h.notify("打开原文失败")
		}
	}()
}

func (h *Header) notify(message string) {
	if h.callbacks.Notify != nil {
		h.callbacks.Notify("error", message)
	}
}

func (h *Header) scaled(theme *ui.Theme, size float32) float32 {
	return size * theme.FontSize / textSize
}

func selectedFeedTitle(snapshot appstatus.Snapshot, feedID int64) string {
	for _, feed := range snapshot.Feeds {
		if feed.ID == feedID {
			return feed.Title
		}
	}
	return ""
}
