package gui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestNewImageViewerContentDecodesPNG(t *testing.T) {
	data, err := testViewerPNG()
	if err != nil {
		t.Fatal(err)
	}

	source := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	content, err := newImageViewerContent(data, source)
	if err != nil {
		t.Fatal(err)
	}
	if content.image == nil || content.url != source {
		t.Fatalf("content = %+v; want a native image and its data URL", content)
	}
}

func TestNewImageViewerContentFallsBackForAVIF(t *testing.T) {
	source := "data:image/avif;base64," + base64.StdEncoding.EncodeToString(testAVIFData())
	content, err := newImageViewerContent(testAVIFData(), source)
	if err != nil {
		t.Fatal(err)
	}
	if content.image != nil || content.url != source {
		t.Fatalf("content = %+v; want a WebView fallback", content)
	}
}

func TestImageViewerClipboardPNGKeepsPNGBytes(t *testing.T) {
	data, err := testViewerPNG()
	if err != nil {
		t.Fatal(err)
	}
	content, err := newImageViewerContent(data, "")
	if err != nil {
		t.Fatal(err)
	}

	got, err := content.clipboardPNG()
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("clipboardPNG() copied %d bytes, error %v; want the original PNG", len(got), err)
	}
}

func TestImageViewerClipboardPNGConvertsJPEG(t *testing.T) {
	var source bytes.Buffer
	if err := jpeg.Encode(&source, image.NewRGBA(image.Rect(0, 0, 2, 1)), nil); err != nil {
		t.Fatal(err)
	}
	content, err := newImageViewerContent(source.Bytes(), "")
	if err != nil {
		t.Fatal(err)
	}

	data, err := content.clipboardPNG()
	if err != nil {
		t.Fatal(err)
	}
	pngData, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if bounds := pngData.Bounds(); bounds.Dx() != 2 || bounds.Dy() != 1 {
		t.Fatalf("copied PNG bounds = %v; want 2x1", bounds)
	}
}

func TestImageViewerOffersCopyContextMenu(t *testing.T) {
	data, err := testViewerPNG()
	if err != nil {
		t.Fatal(err)
	}
	content, err := newImageViewerContent(data, "")
	if err != nil {
		t.Fatal(err)
	}

	tester := ui.NewTester(func(c *ui.Context) { imageViewerView(c, &content, nil) }, 2, 1)
	tester.RightClickAt(1, 0.5)
	if got := tester.Menu(); !slices.Equal(got, []string{"复制图片", "下载图片"}) {
		t.Fatalf("image viewer context menu = %q; want [复制图片 下载图片]", got)
	}
}

func TestNewImageViewerContentRejectsUndecodableData(t *testing.T) {
	if _, err := newImageViewerContent([]byte("not an image"), ""); err == nil {
		t.Fatal("newImageViewerContent accepted invalid image bytes")
	}
}

func TestImageDownloadNameUsesTimestampAndActualFormat(t *testing.T) {
	before := time.Now().Unix()
	got := imageDownloadName(".jpg")
	after := time.Now().Unix()
	if got != fmt.Sprintf("%d.jpg", before) && got != fmt.Sprintf("%d.jpg", after) {
		t.Fatalf("imageDownloadName(.jpg) = %q, want a Unix timestamp between %d and %d", got, before, after)
	}
}

func TestImageTitleTrimsAndLimitsAltText(t *testing.T) {
	if got := imageTitle("  spaced   title  "); got != "spaced title" {
		t.Fatalf("imageTitle() = %q", got)
	}
	if got := imageTitle(""); got != "图片" {
		t.Fatalf("empty imageTitle() = %q", got)
	}

	long := strings.Repeat("文", imageTitleRunes+10)
	got := []rune(imageTitle(long))
	if len(got) != imageTitleRunes+1 || string(got[:imageTitleRunes]) != strings.Repeat("文", imageTitleRunes) ||
		string(got[imageTitleRunes]) != "…" {
		t.Fatalf("long imageTitle has %d runes", len(got))
	}
}

func testViewerPNG() ([]byte, error) {
	var output bytes.Buffer
	if err := png.Encode(&output, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func testAVIFData() []byte {
	data := make([]byte, 20)
	data[0] = 0
	data[1] = 0
	data[2] = 0
	data[3] = 20
	copy(data[4:8], "ftyp")
	copy(data[8:12], "avif")
	copy(data[16:20], "miaf")
	return data
}
