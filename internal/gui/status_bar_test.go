package gui

import (
	"testing"

	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/internal/status"
)

func TestStatusBarVersionOpensAbout(t *testing.T) {
	model := appstatus.New("26.11.8")
	opened := false
	status := NewStatusBar(model, func() {
		opened = true
	})

	tester := ui.NewTester(status.View, 640, 60)
	if err := tester.Click("v26.11.8"); err != nil {
		t.Fatal(err)
	}
	if !opened {
		t.Fatal("clicking the version did not open the About window")
	}
}
