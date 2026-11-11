//go:build windows

package window

import (
	"testing"

	"github.com/egoist/mygo"
)

func TestCenteredWindowPositionUsesParentDisplay(t *testing.T) {
	parent := mygo.Rectangle{X: 2300, Y: 100, Width: 1200, Height: 800}
	child := mygo.Rectangle{Width: 960, Height: 640}
	workArea := mygo.Rectangle{X: 1920, Y: 0, Width: 1920, Height: 1040}

	x, y := centeredWindowPosition(parent, child, workArea)
	if x != 2420 || y != 180 {
		t.Fatalf("centeredWindowPosition() = (%d, %d); want (2420, 180)", x, y)
	}
}

func TestCenteredWindowPositionClampsToWorkArea(t *testing.T) {
	parent := mygo.Rectangle{X: 3000, Y: 800, Width: 1200, Height: 800}
	child := mygo.Rectangle{Width: 960, Height: 640}
	workArea := mygo.Rectangle{X: 1920, Y: 0, Width: 1920, Height: 1040}

	x, y := centeredWindowPosition(parent, child, workArea)
	if x != 2880 || y != 400 {
		t.Fatalf("centeredWindowPosition() = (%d, %d); want (2880, 400)", x, y)
	}
}

func TestCenteredWindowPositionSupportsNegativeCoordinates(t *testing.T) {
	parent := mygo.Rectangle{X: -1500, Y: 200, Width: 1200, Height: 800}
	child := mygo.Rectangle{Width: 960, Height: 640}
	workArea := mygo.Rectangle{X: -1920, Y: 0, Width: 1920, Height: 1040}

	x, y := centeredWindowPosition(parent, child, workArea)
	if x != -1380 || y != 280 {
		t.Fatalf("centeredWindowPosition() = (%d, %d); want (-1380, 280)", x, y)
	}
}
