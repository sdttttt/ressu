package gui

import (
	"fmt"
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
		return imageViewerContent{image: image, url: viewerURL}, nil
	}
	if media.IsAVIF(data) {
		return imageViewerContent{url: viewerURL}, nil
	}
	return imageViewerContent{}, fmt.Errorf("could not decode the image: %w", decodeErr)
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
		options.Content = ui.View(func(c *ui.Context) {
			ui.Image(c, content.image).Fill().Fit(ui.Contain)
		})
	} else {
		options.URL = content.url
	}
	mygo.NewWindow(options)
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
