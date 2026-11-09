package gui

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"math"
	"strings"

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
func OpenImageViewer(parent *mygo.Window, data []byte, fallbackURL string, alt string) error {
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
		return imageViewerContent{url: viewerURL}, nil
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
func openImageViewer(parent *mygo.Window, content imageViewerContent, alt string) {
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
		options.Content = ui.View(func(c *ui.Context) { imageViewerView(c, &content) })
	} else {
		options.URL = content.url
	}
	mygo.NewWindow(options)
}

// imageViewerView draws the image and its native copy action.
func imageViewerView(c *ui.Context, content *imageViewerContent) {
	ui.Image(c, content.image).Fill().Fit(ui.Contain).ContextMenu(func(m *ui.Menu) {
		item := m.Item("复制图片")
		pngData, err := content.clipboardPNG()
		item.Disabled(err != nil)
		if item.Chosen() && err == nil {
			_ = mygo.Clipboard.WriteImage(pngData)
		}
	})
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
