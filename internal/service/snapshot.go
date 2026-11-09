package service

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"strings"
)

// maxSnapshotTiles bounds how many visible-page captures make one snapshot.
const maxSnapshotTiles = 32

// maxSnapshotSide bounds the width and height of the stitched PNG.
const maxSnapshotSide = 16384

// maxSnapshotTileBytes bounds a single capture handed over by the page.
const maxSnapshotTileBytes = 32 << 20

// SnapshotCaptureRequest describes the part of the page a capture shows.
type SnapshotCaptureRequest struct {
	Offset   int64            `json:"offset"`
	Geometry SnapshotGeometry `json:"geometry"`
}

// SnapshotGeometry is the page layout shared by every capture.
type SnapshotGeometry struct {
	ViewportWidth  int `json:"viewportWidth"`
	ViewportHeight int `json:"viewportHeight"`
	CardWidth      int `json:"cardWidth"`
	ContentHeight  int `json:"contentHeight"`
}

// SnapshotTile is one capture and the scroll position it was taken at.
type SnapshotTile struct {
	Offset   int64            `json:"offset"`
	PNG      []byte           `json:"png"`
	Geometry SnapshotGeometry `json:"geometry"`
}

// validate rejects captures whose page geometry cannot be trusted.
func (r SnapshotCaptureRequest) validate() error {
	if r.Offset < 0 || r.Geometry.ContentHeight <= 0 || r.Offset >= int64(r.Geometry.ContentHeight) {
		return fmt.Errorf("invalid snapshot scroll position")
	}
	if err := r.Geometry.validate(); err != nil {
		return err
	}
	return nil
}

// validate rejects geometry whose page measurements cannot be trusted.
func (g SnapshotGeometry) validate() error {
	if g.ViewportWidth <= 0 || g.ViewportHeight <= 0 || g.CardWidth <= 0 || g.ContentHeight <= 0 {
		return fmt.Errorf("invalid snapshot page size")
	}
	if g.CardWidth > g.ViewportWidth {
		return fmt.Errorf("the snapshot card does not fit the window")
	}
	return nil
}

// tileCount is how many overlapping viewport captures the page should take.
func (r SnapshotCaptureRequest) tileCount() int {
	viewportHeight := r.Geometry.ViewportHeight
	if r.Geometry.ContentHeight <= viewportHeight || viewportHeight <= 1 {
		return 1
	}
	remaining := r.Geometry.ContentHeight - viewportHeight
	step := viewportHeight - 1
	return 1 + (remaining+step-1)/step
}

// geometryOf returns the geometry saved with the first tile.
func geometryOf(tiles []SnapshotTile) SnapshotGeometry {
	if len(tiles) == 0 {
		return SnapshotGeometry{}
	}
	return tiles[0].Geometry
}

// maxSnapshotImageBytes bounds a finished snapshot handed back by the page.
const maxSnapshotImageBytes = 256 << 20

// validateSnapshotImage rejects data the clipboard cannot reasonably accept.
func validateSnapshotImage(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("the snapshot is empty")
	}
	if len(data) > maxSnapshotImageBytes {
		return fmt.Errorf("the snapshot is too large")
	}
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("the snapshot is not a PNG image")
	}
	if config.Width > maxSnapshotSide || config.Height > maxSnapshotSide {
		return fmt.Errorf("the snapshot is too large")
	}
	return nil
}

// stitchSnapshot crops each visible-page capture to the snapshot card and
// places it at the physical scroll position it represents.
func stitchSnapshot(tiles []SnapshotTile, geometry SnapshotGeometry) ([]byte, error) {
	if len(tiles) == 0 {
		return nil, fmt.Errorf("the snapshot has no captures")
	}
	if len(tiles) > maxSnapshotTiles {
		return nil, fmt.Errorf("the article is too long for one snapshot")
	}
	if tiles[0].Offset != 0 {
		return nil, fmt.Errorf("snapshot captures are out of order")
	}
	if err := geometry.validate(); err != nil {
		return nil, err
	}

	var (
		scale          float64
		width, height  int
		coveredBottom  int
		previousOffset int64
		previousWidth  int
		previousHeight int
	)
	for i, tile := range tiles {
		if tile.PNG == nil {
			return nil, fmt.Errorf("snapshot capture %d is empty", i+1)
		}
		if len(tile.PNG) > maxSnapshotTileBytes {
			return nil, fmt.Errorf("snapshot capture %d is too large", i+1)
		}
		if tile.Offset < 0 || tile.Offset >= int64(geometry.ContentHeight) {
			return nil, fmt.Errorf("invalid snapshot scroll position")
		}
		if i > 0 && tile.Offset <= previousOffset {
			return nil, fmt.Errorf("snapshot captures are out of order")
		}
		if tile.Geometry != geometry {
			return nil, fmt.Errorf("the article changed while taking the snapshot")
		}
		previousOffset = tile.Offset

		config, err := png.DecodeConfig(bytes.NewReader(tile.PNG))
		if err != nil {
			return nil, fmt.Errorf("snapshot capture %d is not a PNG image", i+1)
		}
		if config.Width > maxSnapshotSide || config.Height > maxSnapshotSide {
			return nil, fmt.Errorf("snapshot capture %d is too large", i+1)
		}
		if i == 0 {
			scale = float64(config.Width) / float64(geometry.ViewportWidth)
			width = roundInt(float64(geometry.CardWidth) * scale)
			height = roundInt(float64(geometry.ContentHeight) * scale)
			if width <= 0 || height <= 0 || width > maxSnapshotSide || height > maxSnapshotSide {
				return nil, fmt.Errorf("the article is too large for one snapshot")
			}
		} else if config.Width != previousWidth || config.Height != previousHeight {
			return nil, fmt.Errorf("the window changed while taking the snapshot")
		}
		previousWidth, previousHeight = config.Width, config.Height
		if config.Width < width {
			return nil, fmt.Errorf("the snapshot card does not fit the capture")
		}
	}

	output := image.NewNRGBA(image.Rect(0, 0, width, height))
	for i, tile := range tiles {
		capture, err := png.Decode(bytes.NewReader(tile.PNG))
		if err != nil {
			return nil, fmt.Errorf("snapshot capture %d is not a PNG image", i+1)
		}

		sourceY := roundInt(float64(tile.Offset) * scale)
		sourceHeight := capture.Bounds().Dy()
		if sourceY >= height {
			return nil, fmt.Errorf("snapshot capture %d does not show its scroll position", i+1)
		}
		drawHeight := min(sourceHeight, height-sourceY)
		if drawHeight <= 0 {
			return nil, fmt.Errorf("snapshot capture %d does not fit the article", i+1)
		}
		if sourceY > coveredBottom {
			return nil, fmt.Errorf("the snapshot captures leave a gap")
		}
		draw.Draw(output, image.Rect(0, sourceY, width, sourceY+drawHeight), capture, image.Pt(0, 0), draw.Src)
		coveredBottom = max(coveredBottom, sourceY+drawHeight)
	}
	if coveredBottom < height {
		return nil, fmt.Errorf("the article changed while taking the snapshot")
	}

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, output); err != nil {
		return nil, fmt.Errorf("could not encode the snapshot: %w", err)
	}
	return encoded.Bytes(), nil
}

// roundInt rounds half away from zero; every value here is non-negative.
func roundInt(value float64) int {
	return int(math.Round(value))
}

// maxFileNameRunes keeps the default file name well under every file
// system's name limit.
const maxFileNameRunes = 120

// snapshotFileName turns a post title into a safe PNG file name, falling
// back to a fixed name when nothing usable remains.
func snapshotFileName(title string) string {
	name := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r < 0x20 {
			return -1
		}
		return r
	}, title)
	name = strings.TrimSpace(name)
	if runes := []rune(name); len(runes) > maxFileNameRunes {
		name = string(runes[:maxFileNameRunes])
	}
	name = strings.TrimRight(name, " .")
	if strings.HasSuffix(strings.ToLower(name), ".png") {
		name = name[:len(name)-4]
	}
	if name == "" {
		name = "文章快照"
	}
	return name + ".png"
}
