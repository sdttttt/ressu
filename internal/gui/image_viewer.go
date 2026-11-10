package gui

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/media"
)

const (
	// imageViewerWidth is the initial content width of a viewer window.
	imageViewerWidth = 960

	// imageViewerHeight is the initial content height of a viewer window.
	imageViewerHeight = 640

	// imageTitleRunes keeps article-controlled alt text from creating an
	// unusably long native window title.
	imageTitleRunes = 80
)

// imageViewerContent holds either a decoded native image or a safe URL for
// formats the native UI cannot decode.
type imageViewerContent struct {
	image ui.ImageSource
	url   string

	data           []byte
	clipboardBytes []byte
	clipboardErr   error
	clipboardReady bool
}

// OpenImageViewer renders validated image bytes in a child window. AVIF falls
// back to a child WebView because the native image decoder cannot read it.
func OpenImageViewer(parent *Window, data []byte, fallbackURL string, alt string) error {
	content, err := newImageViewerContent(data, fallbackURL)
	if err != nil {
		return err
	}
	openImageViewer(parent, content, alt)
	return nil
}

// newImageViewerContent resolves safe viewer bytes and decodes the native
// image when possible.
func newImageViewerContent(data []byte, viewerURL string) (imageViewerContent, error) {
	image, decodeErr := ui.DecodeImage(data)
	if decodeErr == nil {
		return imageViewerContent{image: image, url: viewerURL, data: data}, nil
	}
	if media.IsAVIF(data) {
		return imageViewerContent{url: viewerURL, data: data}, nil
	}
	return imageViewerContent{}, fmt.Errorf("could not decode the image: %w", decodeErr)
}

// clipboardPNG returns the image as the PNG bytes system clipboards accept.
// The result is prepared lazily and kept for the lifetime of the viewer.
func (content *imageViewerContent) clipboardPNG() ([]byte, error) {
	if content.clipboardReady {
		return content.clipboardBytes, content.clipboardErr
	}

	if bytes.HasPrefix(content.data, []byte("\x89PNG\r\n\x1a\n")) {
		content.clipboardBytes, content.clipboardErr = content.data, nil
	} else {
		content.clipboardBytes, content.clipboardErr = renderClipboardPNG(content.image)
	}
	content.clipboardReady = true
	return content.clipboardBytes, content.clipboardErr
}

// renderClipboardPNG rasterizes a native image at its intrinsic size.
func renderClipboardPNG(source ui.ImageSource) ([]byte, error) {
	width, height := ui.ImageSize(source)
	pixelWidth := int(math.Ceil(float64(width)))
	pixelHeight := int(math.Ceil(float64(height)))
	if pixelWidth <= 0 || pixelHeight <= 0 {
		return nil, errors.New("could not copy an image without dimensions")
	}

	pixels := ui.Render(func(c *ui.Context) {
		ui.Image(c, source).Fill().Fit(ui.Contain)
	}, pixelWidth, pixelHeight, 1)

	var output bytes.Buffer
	if err := png.Encode(&output, pixels); err != nil {
		return nil, fmt.Errorf("could not encode the copied image: %w", err)
	}
	return output.Bytes(), nil
}

// openImageViewer creates the native viewer window for prepared content.
func openImageViewer(parent *Window, content imageViewerContent, alt string) {
	options := mygo.WindowOptions{
		Title:           imageTitle(alt),
		Width:           imageViewerWidth,
		Height:          imageViewerHeight,
		UseContentSize:  true,
		MinWidth:        320,
		MinHeight:       240,
		Parent:          parent,
		BackgroundColor: "#000000",
	}

	if content.image != nil {
		var viewer *Window
		options.Content = ui.View(func(c *ui.Context) { imageViewerView(c, &content, viewer) })
		viewer = mygo.NewWindow(options)
		return
	} else {
		options.URL = content.url
	}
	mygo.NewWindow(options)
}

// imageViewerView draws the image and its native viewer actions.
func imageViewerView(c *ui.Context, content *imageViewerContent, viewer *Window) {
	ui.Image(c, content.image).Fill().Fit(ui.Contain).ContextMenu(func(m *ui.Menu) {
		copyItem := m.Item("复制图片")
		pngData, err := content.clipboardPNG()
		copyItem.Disabled(err != nil)
		if copyItem.Chosen() && err == nil {
			_ = mygo.Clipboard.WriteImage(pngData)
		}

		downloadItem := m.Item("下载图片")
		downloadItem.Disabled(viewer == nil || len(content.data) == 0)
		if downloadItem.Chosen() {
			downloadImageViewerImage(viewer, content)
		}
	})
}

// downloadImageViewerImage asks where to keep the image and writes its
// validated bytes to the selected path.
func downloadImageViewerImage(viewer *Window, content *imageViewerContent) {
	extension := imageExtension(content.data)
	savePath, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      viewer,
		Title:       "下载图片",
		DefaultPath: imageDownloadName(extension),
		ButtonLabel: "保存",
		Filters:     []mygo.FileFilter{{Name: "图片", Extensions: []string{strings.TrimPrefix(extension, ".")}}},
	})
	if err != nil {
		showImageSaveError(viewer, err)
		return
	}
	if savePath == "" {
		return
	}
	if err := os.WriteFile(savePath, content.data, 0o644); err != nil {
		showImageSaveError(viewer, fmt.Errorf("could not save the image: %w", err))
	}
}

// showImageSaveError reports a failed download without losing the viewer.
func showImageSaveError(viewer *Window, err error) {
	_, _ = mygo.Dialog.Message(mygo.MessageOptions{
		Parent:  viewer,
		Type:    mygo.MessageError,
		Title:   "下载失败",
		Message: "无法保存图片",
		Detail:  err.Error(),
	})
}

// imageExtension chooses a download extension from validated image bytes.
func imageExtension(data []byte) string {
	switch media.ContentType(data) {
	case "image/avif":
		return ".avif"
	case "image/bmp":
		return ".bmp"
	case "image/gif":
		return ".gif"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	default:
		return ".png"
	}
}

// imageDownloadName creates a default file name from the current Unix time
// and the image's actual extension.
func imageDownloadName(extension string) string {
	return fmt.Sprintf("%d%s", time.Now().Unix(), extension)
}

// imageTitle chooses a concise, article-controlled window title.
func imageTitle(alt string) string {
	title := strings.Join(strings.Fields(strings.TrimSpace(alt)), " ")
	if title == "" {
		return "图片"
	}

	runes := []rune(title)
	if len(runes) <= imageTitleRunes {
		return title
	}
	return string(runes[:imageTitleRunes]) + "…"
}
