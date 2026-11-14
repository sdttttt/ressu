// Package addfeed renders the native subscription discovery, editing, and
// OPML import window.
package addfeed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/store"
	guitheme "github.com/sdttttt/ressu/mygoui/theme"
	appwindow "github.com/sdttttt/ressu/mygoui/window"
)

const (
	// addFeedWidth is the initial content width of the add-feed window.
	addFeedWidth = 460

	// addFeedHeight is the initial content height of the add-feed window.
	addFeedHeight = 480

	// addFeedMinHeight is the minimum content height of the add-feed window.
	addFeedMinHeight = 300

	// addFeedPreviewExtraHeight is added when a feed preview is shown.
	addFeedPreviewExtraHeight = 100
)

var addFeedWindow struct {
	sync.Mutex
	window *mygo.Window
}

// Preview is the native subscription preview.
type Preview struct {
	Title       string `json:"title"`
	SiteURL     string `json:"siteUrl"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	Language    string `json:"language"`
}

// Summary reports the result of importing a subscription list.
type Summary struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors"`
}

// Callbacks are the operations a native add-feed surface may invoke.
type Callbacks struct {
	Appearance func() store.Prefs
	Test       func(context.Context, string, int64) (Preview, error)
	Add        func(context.Context, string, string, string) (store.Feed, error)
	Save       func(context.Context, int64, string, string, string) (store.Feed, error)
	ImportOPML func(parent *mygo.Window) (Summary, error)
}

// Open opens the single native subscription window for adding a feed.
func Open(parent *mygo.Window, callbacks Callbacks) {
	open(parent, nil, callbacks)
}

// OpenEdit opens the same window pre-filled for editing an existing feed.
func OpenEdit(parent *mygo.Window, feed store.Feed, callbacks Callbacks) {
	open(parent, &feed, callbacks)
}

// open opens the single native subscription window, optionally in edit mode.
func open(parent *mygo.Window, edit *store.Feed, callbacks Callbacks) {
	addFeedWindow.Lock()
	if window := addFeedWindow.window; window != nil && !window.IsDestroyed() {
		addFeedWindow.Unlock()
		window.Show()
		window.Focus()
		return
	}
	addFeedWindow.Unlock()

	prefs := callbacks.Appearance()
	view := &addFeedView{
		callbacks:       callbacks,
		prefs:           prefs,
		edit:            edit,
		remoteResources: prefs.RemoteResources,
	}
	if edit != nil {
		view.url = edit.URL
		view.testedURL = edit.URL
		if edit.RemoteResources != "" {
			view.remoteResources = edit.RemoteResources
		}
		if edit.CustomTitle {
			view.title = edit.Title
			view.titleTouched = true
		}
	}
	title := "添加订阅源"
	if edit != nil {
		title = "编辑订阅源"
	}
	options := mygo.WindowOptions{
		Title:           title,
		Width:           addFeedWidth,
		Height:          addFeedHeight,
		UseContentSize:  true,
		MinWidth:        400,
		MinHeight:       addFeedMinHeight,
		Parent:          parent,
		StateKey:        "add-feed",
		BackgroundColor: "light-dark(#ffffff, #101012)",
		Content:         ui.View(func(c *ui.Context) { view.draw(c) }),
	}
	window := appwindow.New(parent, options)
	view.window = window

	addFeedWindow.Lock()
	addFeedWindow.window = window
	addFeedWindow.Unlock()
	window.OnClosed(func() {
		addFeedWindow.Lock()
		if addFeedWindow.window == window {
			addFeedWindow.window = nil
		}
		addFeedWindow.Unlock()
	})
}

// addFeedView keeps the subscription draft, preview, and operation state.
type addFeedView struct {
	callbacks       Callbacks
	prefs           store.Prefs
	edit            *store.Feed
	window          *mygo.Window
	url             string
	title           string
	remoteResources string
	testedURL       string
	titleTouched    bool
	lastTitle       string
	testing         bool
	pending         bool
	importing       bool
	err             string
	preview         *Preview
	previewDelta    int
}

// draw renders one frame of the subscription form.
func (s *addFeedView) draw(c *ui.Context) {
	theme := guitheme.New(c, s.prefs.Theme, s.prefs.FontSize)

	ui.Themed(c, theme).Fill().Children(func() {
		ui.Scroll(c).Grow(1).MinHeight(0).Padding(20, 24).Children(func() {
			ui.Form(c, func() {
				ui.Field(c, "订阅地址", func() {
					if ui.TextInput(c, &s.url).
						Placeholder("https://example.com/feed.xml").Changed() {
						s.invalidateTest()
					}
				})
				ui.Field(c, "订阅名称", func() {
					input := ui.TextInput(c, &s.title)
					if s.titleTouched {
						input.Placeholder(s.previewTitle())
					} else {
						input.Placeholder("自动使用订阅源名称")
					}
					if s.title != "" && s.title != s.lastTitle {
						s.titleTouched = true
					}
					s.lastTitle = s.title
				}).Description("留空时使用订阅源名称。")
				ui.Field(c, "远程资源访问方式", func() {
					ui.RadioGroup(c, func() {
						ui.Radio(c, &s.remoteResources,
							store.ResourceAccessGoProxy, "由 Ressu 代理并缓存")
						ui.Radio(c, &s.remoteResources,
							store.ResourceAccessWebview, "由 WebView 直接加载")
					})
				}).Description("只影响这个订阅源的正文图片加载。")
			})

			ui.Column(c).Padding(16, 0).Gap(8).Children(func() {
				testLabel := "测试地址"
				if s.testing {
					testLabel = "正在测试…"
				}
				if ui.Button(c, testLabel).Disabled(s.testing || s.pending || s.importing).Clicked() {
					s.startTest()
				}
				addLabel := "添加订阅"
				if s.edit != nil {
					addLabel = "保存修改"
				}
				if s.pending {
					if s.edit != nil {
						addLabel = "正在保存…"
					} else {
						addLabel = "正在添加…"
					}
				}
				feedURL := strings.TrimSpace(s.url)
				addDisabled := s.pending || s.testing || s.importing || feedURL == ""
				if feedURL != s.testedURL && (s.edit == nil || feedURL != s.edit.URL) {
					addDisabled = true
				}
				if ui.Button(c, addLabel).Disabled(addDisabled).Clicked() {
					s.startAdd()
				}
			})

			ui.Column(c).Gap(8).Children(func() {
				if s.preview != nil {
					s.drawPreview(c, theme)
				}
				if s.err != "" {
					ui.Text(c, s.err).TextColor(theme.Danger)
				}
			})
		})

		ui.Box(c).Height(1).Shrink(0).Background(theme.Border)
		ui.Row(c).Shrink(0).Padding(6, 10).Gap(6).AlignItems(ui.Center).Children(func() {
			switch {
			case s.err != "":
				ui.Text(c, s.err).TextColor(theme.Danger).Grow(1)
			case s.importing:
				ui.Spinner(c)
				ui.Text(c, "正在导入订阅列表…").TextColor(theme.TextMuted).Grow(1)
			default:
				ui.Text(c, "支持 RSS、Atom 和 JSON Feed").TextColor(theme.TextMuted).Grow(1)
			}
			if s.edit == nil {
				importLabel := "导入 OPML…"
				if s.importing {
					importLabel = "正在导入…"
				}
				if ui.Button(c, importLabel).Disabled(s.pending || s.testing || s.importing).Clicked() {
					s.startImport()
				}
			}
		})
	})
}

// drawPreview renders the validated subscription metadata.
func (s *addFeedView) drawPreview(c *ui.Context, theme *ui.Theme) {
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, "订阅预览").TextColor(theme.TextMuted).FontSize(12)
		ui.Text(c, s.previewTitle())
		if s.preview.SiteURL != "" {
			ui.Text(c, s.preview.SiteURL).TextColor(theme.TextMuted)
		}
		if s.preview.Description != "" {
			ui.Text(c, s.preview.Description)
		}
		if s.preview.Language != "" {
			ui.Textf(c, "语言：%s", s.preview.Language).
				TextColor(theme.TextMuted)
		}
	})
}

// previewTitle chooses the user title, preview title, or URL.
func (s *addFeedView) previewTitle() string {
	if s.title != "" {
		return s.title
	}
	if s.preview != nil && s.preview.Title != "" {
		return s.preview.Title
	}
	return strings.TrimSpace(s.url)
}

// startTest validates the address without creating a subscription.
func (s *addFeedView) startTest() {
	feedURL := strings.TrimSpace(s.url)
	if feedURL == "" || !s.beginBusy(&s.testing, "测试中不能同时执行其他操作") {
		return
	}
	s.clearError()
	s.invalidate()

	go func() {
		var excludeID int64
		if s.edit != nil {
			excludeID = s.edit.ID
		}
		preview, err := s.callbacks.Test(context.Background(), feedURL, excludeID)
		s.complete(func() {
			s.testing = false
			if err != nil {
				s.fail(err)
				return
			}
			s.testedURL = feedURL
			s.preview = &preview
			if !s.titleTouched {
				s.title = ""
			}
			s.resizeForPreview(true)
		})
	}()
}

// startAdd creates the subscription, or saves the edited one, and closes the
// window on success.
func (s *addFeedView) startAdd() {
	feedURL := strings.TrimSpace(s.url)
	title := strings.TrimSpace(s.title)
	tested := feedURL != "" && (feedURL == s.testedURL || (s.edit != nil && feedURL == s.edit.URL))
	if !tested {
		s.err = "请先测试订阅地址"
		if s.edit != nil {
			s.err = "地址已更改，请先测试订阅地址"
		}
		s.invalidate()
		return
	}
	if !s.beginBusy(&s.pending, "处理中不能同时执行其他操作") {
		return
	}
	s.clearError()
	s.invalidate()

	go func() {
		var err error
		if s.edit != nil {
			_, err = s.callbacks.Save(context.Background(), s.edit.ID, title, feedURL, s.remoteResources)
		} else {
			_, err = s.callbacks.Add(context.Background(), title, feedURL, s.remoteResources)
		}
		window := s.window
		s.complete(func() {
			s.pending = false
			if err != nil {
				s.fail(err)
				return
			}
			if s.edit == nil {
				s.reset()
			}
		})
		if err == nil && window != nil && !window.IsDestroyed() {
			window.Close()
		}
	}()
}

// startImport opens the OPML picker and imports the selected file.
func (s *addFeedView) startImport() {
	if !s.beginBusy(&s.importing, "导入中不能同时执行其他操作") {
		return
	}
	s.clearError()
	s.invalidate()
	window := s.window

	go func() {
		summary, err := s.callbacks.ImportOPML(window)
		s.complete(func() {
			s.importing = false
			if err != nil {
				s.fail(err)
				return
			}
			switch {
			case summary.Imported == 0 && summary.Skipped == 0 && summary.Failed == 0:
				return
			case summary.Failed > 0:
				message := fmt.Sprintf("导入 %d 个，跳过 %d 个，失败 %d 个",
					summary.Imported, summary.Skipped, summary.Failed)
				if len(summary.Errors) > 0 {
					message += "：" + summary.Errors[0]
				}
				s.fail(errors.New(message))
			case summary.Skipped > 0:
				s.notice(fmt.Sprintf("导入 %d 个，跳过 %d 个",
					summary.Imported, summary.Skipped))
			default:
				s.notice(fmt.Sprintf("已导入 %d 个订阅源", summary.Imported))
			}
		})
	}()
}

// beginBusy guards against overlapping native operations.
func (s *addFeedView) beginBusy(field *bool, message string) bool {
	if *field || s.testing || s.pending || s.importing {
		s.err = message
		s.invalidate()
		return false
	}
	*field = true
	return true
}

// complete applies an async result on the UI thread.
func (s *addFeedView) complete(update func()) {
	window := s.window
	if window == nil || window.IsDestroyed() {
		return
	}
	window.Update(func() {
		if s.window == window && !window.IsDestroyed() {
			update()
		}
	})
}

// reset clears the completed subscription draft.
func (s *addFeedView) reset() {
	s.url = ""
	s.title = ""
	s.lastTitle = ""
	s.testedURL = ""
	s.titleTouched = false
	s.preview = nil
	s.clearError()
}

// clearError removes the previous operation result.
func (s *addFeedView) clearError() {
	s.err = ""
}

// invalidateTest drops the preview when the tested address changes.
func (s *addFeedView) invalidateTest() {
	if s.testedURL != "" && strings.TrimSpace(s.url) != s.testedURL {
		s.testedURL = ""
		s.preview = nil
		s.resizeForPreview(false)
	}
	s.clearError()
}

// resizeForPreview adjusts the content height once per preview state change.
func (s *addFeedView) resizeForPreview(visible bool) {
	window := s.window
	if window == nil || window.IsDestroyed() {
		return
	}

	delta := 0
	if visible {
		if s.previewDelta != 0 {
			return
		}
		delta = addFeedPreviewExtraHeight
	} else {
		if s.previewDelta == 0 {
			return
		}
		delta = -s.previewDelta
	}

	width, height := window.ContentSize()
	s.previewDelta += delta
	window.SetContentSize(width, max(height+delta, addFeedMinHeight))
}

// notice shows a short import confirmation.
func (s *addFeedView) notice(message string) {
	s.err = message
}

// fail records an error message for the next frame.
func (s *addFeedView) fail(err error) {
	s.err = "操作失败：" + err.Error()
}

// invalidate requests a redraw when the window still exists.
func (s *addFeedView) invalidate() {
	if window := s.window; window != nil && !window.IsDestroyed() {
		window.Invalidate()
	}
}
