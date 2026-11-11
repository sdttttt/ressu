package gui

import (
	"errors"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
	"github.com/sdttttt/ressu/internal/store"
)

const (
	// snapshotViewerWidth is the initial content width of the preview window.
	snapshotViewerWidth = 960

	// snapshotViewerHeight is the initial content height of the preview window.
	snapshotViewerHeight = 640

	// snapshotTitleRunes keeps article-controlled titles from filling the
	// snapshot toolbar and window title.
	snapshotTitleRunes = 80
)

// SnapshotViewerCallbacks handles actions for a generated article snapshot.
type SnapshotViewerCallbacks struct {
	Save   func(parent *Window, data []byte, title string) (string, error)
	Copy   func(data []byte) error
	Notify func(level, message string)
}

// snapshotViewerContent draws a stitched snapshot and its native actions.
type snapshotViewerContent struct {
	image     imageViewerContent
	title     string
	callbacks SnapshotViewerCallbacks
	parent    *Window
	window    *Window
	busy      bool
}

// OpenSnapshotViewer opens a native window showing a generated snapshot.
func OpenSnapshotViewer(
	parent *Window,
	data []byte,
	title string,
	callbacks SnapshotViewerCallbacks,
) error {
	content, err := newSnapshotViewerContent(data, title, parent, callbacks)
	if err != nil {
		return err
	}

	window := newChildWindow(parent, mygo.WindowOptions{
		Title:           content.title,
		Width:           snapshotViewerWidth,
		Height:          snapshotViewerHeight,
		UseContentSize:  true,
		MinWidth:        640,
		MinHeight:       480,
		Parent:          parent,
		StateKey:        "snapshot-viewer",
		BackgroundColor: "#ffffff",
		Content:         ui.View(func(c *ui.Context) { content.draw(c) }),
	})
	content.window = window
	return nil
}

// newSnapshotViewerContent validates and decodes a stitched PNG.
func newSnapshotViewerContent(
	data []byte,
	title string,
	parent *Window,
	callbacks SnapshotViewerCallbacks,
) (*snapshotViewerContent, error) {
	image, err := newImageViewerContent(data, "")
	if err != nil {
		return nil, err
	}
	if image.image == nil {
		return nil, errors.New("the snapshot is not a native image")
	}
	return &snapshotViewerContent{
		image:     image,
		title:     snapshotTitle(title),
		callbacks: callbacks,
		parent:    parent,
	}, nil
}

// draw renders the preview toolbar and scrollable snapshot image.
func (v *snapshotViewerContent) draw(c *ui.Context) {
	theme := guitheme.New(c, store.ThemeLight, store.FontSizeMedium)
	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Background(theme.Background).Children(func() {
			ui.Row(c).Shrink(0).
				Padding(theme.Space(3), theme.Space(4)).
				Gap(theme.Space(3)).
				AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, v.title).
						Grow(1).
						MinWidth(0).
						MaxLines(1).
						Wrap().
						Ellipsis("…").
						FontWeight(600).
						TextColor(theme.Text)
					v.drawSave(c)
					v.drawCopy(c)
				})
			ui.Box(c).Shrink(0).FillWidth().Height(1).Background(theme.Border)
			ui.Scroll(c).Grow(1).MinHeight(0).
				Padding(theme.Space(4)).
				Background(theme.Background).
				Children(func() {
					width, height := ui.ImageSize(v.image.image)
					ui.Image(c, v.image.image).
						MaxWidthPercent(100).
						AspectRatio(width / height)
				})
		})
	})
}

// drawSave renders the save action and reports its result.
func (v *snapshotViewerContent) drawSave(c *ui.Context) {
	save := ui.PrimaryButton(c, "保存")
	save.Disabled(v.busy)
	if save.Clicked() && !v.busy && v.callbacks.Save != nil {
		v.busy = true
		path, err := v.callbacks.Save(v.parent, v.image.data, v.title)
		v.busy = false
		if err != nil {
			v.notify("error", "无法保存快照")
			return
		}
		if path == "" {
			return
		}
		v.notify("success", "快照已保存")
		v.notify("info", path)
	}
}

// drawCopy renders the clipboard action and reports its result.
func (v *snapshotViewerContent) drawCopy(c *ui.Context) {
	copy := ui.Button(c, "复制")
	copy.Disabled(v.busy)
	if copy.Clicked() && !v.busy && v.callbacks.Copy != nil {
		v.busy = true
		err := v.callbacks.Copy(v.image.data)
		v.busy = false
		if err != nil {
			v.notify("error", "无法复制快照")
			return
		}
		v.notify("success", "快照已复制到剪贴板")
	}
}

// notify forwards a result to application notifications, falling back to the
// viewer when no queue is available in tests.
func (v *snapshotViewerContent) notify(level, message string) {
	if v.callbacks.Notify == nil {
		return
	}
	v.callbacks.Notify(level, message)
	if v.window != nil && !v.window.IsDestroyed() {
		v.window.Invalidate()
	}
}

// snapshotTitle chooses a concise, article-controlled snapshot title.
func snapshotTitle(title string) string {
	title = strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
	if title == "" {
		return "文章快照"
	}
	if runes := []rune(title); len(runes) > snapshotTitleRunes {
		return string(runes[:snapshotTitleRunes]) + "…"
	}
	return title
}
