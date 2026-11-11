package about

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/sdttttt/ressu/internal/gui/assets"
)

func TestAboutViewRendersMetadata(t *testing.T) {
	view := &aboutView{
		info: Info{
			Version:   "26.11.8",
			BuildDate: "2026-11-10",
			Developer: "sdttttt",
			Email:     "sendaihazuki@foxmail.com",
		},
	}

	tester := ui.NewTester(view.draw, aboutWidth, aboutHeight)
	for _, want := range []string{
		"Ressu",
		"版本 26.11.8",
		"开发者",
		"sdttttt",
		"构建日期",
		"2026-11-10",
		"邮箱",
		"sendaihazuki@foxmail.com",
		"必开花繁盛，乐上加乐，而且欢呼。",
	} {
		if !tester.HasText(want) {
			t.Errorf("About window is missing %q", want)
		}
	}
	if tester.HasText("关闭") {
		t.Error("About window still contains a close button")
	}
}

func TestAboutAvatarDecodes(t *testing.T) {
	if _, err := ui.DecodeBitmap(assets.AboutAvatar); err != nil {
		t.Fatalf("decode About avatar: %v", err)
	}
}
