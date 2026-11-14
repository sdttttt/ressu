package dialog

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestHelpModalShowsShortcutsAndDismisses(t *testing.T) {
	help := NewHelp(func() (string, string) {
		return "system", "small"
	})
	help.Show()
	if !help.Visible() {
		t.Fatal("help was not visible immediately after Show")
	}

	tester := ui.NewTester(help.Draw, 520, 420)
	tester.Frame()
	for _, shortcut := range helpShortcuts {
		if !tester.HasText(shortcut.key) || !tester.HasText(shortcut.description) {
			t.Errorf("help modal is missing %q shortcut or description", shortcut.key)
		}
	}
	if !tester.HasText("快捷键") {
		t.Error("help modal is missing its title")
	}

	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	if help.open {
		t.Error("help modal stayed open after Escape")
	}
	if help.Visible() {
		t.Error("help visibility stayed true after Escape")
	}
}
