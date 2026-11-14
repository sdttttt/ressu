package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// tilePNG encodes a capture whose rows are numbered in their red channel.
// The columns beyond cardWidth stay black, matching the unused page width.
func tilePNG(t *testing.T, viewportWidth, height, startRow int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, viewportWidth, height))
	for y := range height {
		number := uint8(startRow + y)
		for x := range viewportWidth {
			value := color.NRGBA{}
			if x < 16 {
				value = color.NRGBA{R: number, A: 255}
			}
			img.Set(x, y, value)
		}
	}

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestCaptureSnapshotValidatesRequest(t *testing.T) {
	requests := []SnapshotCaptureRequest{
		{Offset: -1, Geometry: geometry(100, 100, 80, 0, 200)},
		{Offset: 200, Geometry: geometry(100, 100, 80, 0, 200)},
		{Offset: 0, Geometry: geometry(0, 100, 80, 0, 200)},
		{Offset: 0, Geometry: geometry(100, 0, 80, 0, 200)},
		{Offset: 0, Geometry: geometry(100, 100, 0, 0, 200)},
		{Offset: 0, Geometry: geometry(100, 100, 101, 0, 200)},
		{Offset: 0, Geometry: geometry(100, 100, 80, 30, 200)},
		{Offset: 0, Geometry: geometry(100, 100, 80, -1, 200)},
		{
			Offset: 0,
			Geometry: geometry(
				100,
				10,
				80,
				0,
				maxSnapshotTiles*10+100,
			),
		},
	}
	for _, request := range requests {
		if err := ValidateSnapshotCapture(request); err == nil {
			t.Errorf("ValidateSnapshotCapture(%+v) should fail", request)
		}
	}
}

func geometry(viewportWidth, viewportHeight, cardWidth, cardLeft, contentHeight int) SnapshotGeometry {
	return SnapshotGeometry{
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
		CardWidth:      cardWidth,
		CardLeft:       cardLeft,
		ContentHeight:  contentHeight,
	}
}

func TestStitchSnapshot(t *testing.T) {
	sharedGeometry := geometry(8, 2, 8, 0, 4)
	tiles := []SnapshotTile{
		{Offset: 0, PNG: tilePNG(t, 16, 4, 0), Geometry: sharedGeometry},
		{Offset: 2, PNG: tilePNG(t, 16, 4, 4), Geometry: sharedGeometry},
	}

	encoded, err := stitchSnapshot(tiles, sharedGeometry)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 16 || bounds.Dy() != 8 {
		t.Fatalf("bounds = %v, want 16x8", bounds)
	}

	// The second capture starts after a full viewport-sized capture. It must
	// still be valid and be drawn at its physical article offset.
	for _, point := range []image.Point{{X: 0, Y: 0}, {X: 15, Y: 7}} {
		want := uint8(point.Y)
		red, _, _, alpha := decoded.At(point.X, point.Y).RGBA()
		if red != uint32(want)*0x101 || alpha != 0xffff {
			t.Errorf("pixel %v red = %d, want red %d", point, red, want)
		}
	}
}

// TestStitchSnapshotCropsCenteredCard covers the capture window layout: the
// card is centered inside a wider viewport, so stitching must crop starting
// at the card's physical offset, not at the capture's left edge.
func TestStitchSnapshotCropsCenteredCard(t *testing.T) {
	geometry := geometry(8, 2, 4, 2, 2)

	// One full-viewport capture at scale 2 (16x4 pixels): the card occupies
	// pixels [4, 12) and is red; the margins on both sides stay black.
	img := image.NewNRGBA(image.Rect(0, 0, 16, 4))
	for x := 4; x < 12; x++ {
		for y := range 4 {
			img.Set(x, y, color.NRGBA{R: 7, A: 255})
		}
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatal(err)
	}

	encoded, err := stitchSnapshot([]SnapshotTile{
		{Offset: 0, PNG: buffer.Bytes(), Geometry: geometry},
	}, geometry)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 8 || decoded.Bounds().Dy() != 4 {
		t.Fatalf("bounds = %v, want 8x4", decoded.Bounds())
	}
	for _, point := range []image.Point{{X: 0, Y: 0}, {X: 7, Y: 3}} {
		red, _, _, alpha := decoded.At(point.X, point.Y).RGBA()
		if red != 7*0x101 || alpha != 0xffff {
			t.Errorf("pixel %v red = %d, want red 7", point, red)
		}
	}
}

func TestStitchSnapshotRejectsBadTiles(t *testing.T) {
	sharedGeometry := geometry(8, 2, 8, 0, 3)
	valid := tilePNG(t, 16, 4, 0)
	cases := []struct {
		name  string
		tiles []SnapshotTile
	}{
		{"empty", nil},
		{"bad png", []SnapshotTile{{Offset: 0, PNG: []byte("broken"), Geometry: sharedGeometry}}},
		{"out of order", []SnapshotTile{
			{Offset: 1, PNG: valid, Geometry: sharedGeometry},
			{Offset: 0, PNG: valid, Geometry: sharedGeometry},
		}},
		{"first offset", []SnapshotTile{
			{Offset: 1, PNG: valid, Geometry: sharedGeometry},
		}},
		{"changed geometry", []SnapshotTile{
			{Offset: 0, PNG: valid, Geometry: sharedGeometry},
			{Offset: 1, PNG: valid, Geometry: geometry(8, 3, 8, 0, 3)},
		}},
		{"changed size", []SnapshotTile{
			{Offset: 0, PNG: valid, Geometry: sharedGeometry},
			{Offset: 1, PNG: tilePNG(t, 15, 4, 2), Geometry: sharedGeometry},
		}},
		{"card outside capture", []SnapshotTile{
			{Offset: 0, PNG: tilePNG(t, 3, 4, 0), Geometry: geometry(4, 2, 2, 2, 3)},
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := stitchSnapshot(testCase.tiles, sharedGeometry); err == nil {
				t.Fatal("stitching should fail")
			}
		})
	}

	var tooMany []SnapshotTile
	for range maxSnapshotTiles + 1 {
		tooMany = append(tooMany, SnapshotTile{Offset: int64(len(tooMany)), PNG: valid, Geometry: sharedGeometry})
	}
	if _, err := stitchSnapshot(tooMany, sharedGeometry); err == nil {
		t.Fatal("too many tiles should fail")
	}
}

func TestSnapshotFileName(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"一次有趣的访谈", "一次有趣的访谈.png"},
		{"  spaces  around  ", "spaces  around.png"},
		{"path/../trick:down", "path..trickdown.png"},
		{"Report.Png", "Report.png"},
		{"already.png", "already.png"},
		{"...", "文章快照.png"},
		{"", "文章快照.png"},
	}
	for _, testCase := range cases {
		if got := SnapshotFileName(testCase.title); got != testCase.want {
			t.Errorf("snapshotFileName(%q) = %q, want %q", testCase.title, got, testCase.want)
		}
	}
}

func TestSnapshotImageValidation(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":  nil,
		"broken": []byte("broken"),
	} {
		if err := ValidateSnapshotImage(data); err == nil {
			t.Errorf("ValidateSnapshotImage(%s) should fail", name)
		}
	}
}
