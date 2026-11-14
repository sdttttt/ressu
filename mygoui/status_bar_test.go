package mygoui

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestStatusBarVersionOpensAbout(t *testing.T) {
	model := app.NewStatusModel("26.11.8")
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
