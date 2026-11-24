package store

import (
	"math"
	"strings"
	"testing"
)

func TestPrefsNormalizeFontFamilies(t *testing.T) {
	if prefs := DefaultPrefs(); prefs.UIFontFamily != FontFamilySystem ||
		prefs.ArticleFontFamily != FontFamilySystem {
		t.Errorf("default font families = %q / %q, want %q",
			prefs.UIFontFamily, prefs.ArticleFontFamily, FontFamilySystem)
	}

	prefs := Prefs{UIFontFamily: FontFamilySerif, ArticleFontFamily: FontFamilyMono}
	prefs.Normalize()
	if prefs.UIFontFamily != FontFamilySerif || prefs.ArticleFontFamily != FontFamilyMono {
		t.Errorf("preset font families = %q / %q", prefs.UIFontFamily, prefs.ArticleFontFamily)
	}

	// Empty values from older databases migrate to the system default.
	prefs = Prefs{UIFontFamily: ""}
	prefs.Normalize()
	if prefs.UIFontFamily != FontFamilySystem {
		t.Errorf("empty font family = %q, want %q", prefs.UIFontFamily, FontFamilySystem)
	}

	// Custom lists survive, trimmed and stripped of CSS-breaking characters.
	prefs = Prefs{ArticleFontFamily: `  "Microsoft YaHei", Arial; body { } <script>  `}
	prefs.Normalize()
	if prefs.ArticleFontFamily != `"Microsoft YaHei", Arial body script` {
		t.Errorf("custom font family = %q", prefs.ArticleFontFamily)
	}

	// Control characters are stripped too.
	prefs = Prefs{ArticleFontFamily: "Aa\x00\x1fBb"}
	prefs.Normalize()
	if prefs.ArticleFontFamily != "AaBb" {
		t.Errorf("control chars in font family = %q", prefs.ArticleFontFamily)
	}

	// A value that sanitizes to empty resets to the default.
	prefs = Prefs{UIFontFamily: ";<>\x00"}
	prefs.Normalize()
	if prefs.UIFontFamily != FontFamilySystem {
		t.Errorf("sanitized-out font family = %q, want %q", prefs.UIFontFamily, FontFamilySystem)
	}
}

func TestPrefsNormalizeFontFamilyLengthCap(t *testing.T) {
	prefs := Prefs{ArticleFontFamily: strings.Repeat("a", MaxFontFamilyLen*2)}
	prefs.Normalize()
	if got := len(prefs.ArticleFontFamily); got != MaxFontFamilyLen {
		t.Errorf("font family length = %d, want %d", got, MaxFontFamilyLen)
	}
}

func TestPrefsNormalizeArticleMetrics(t *testing.T) {
	if prefs := DefaultPrefs(); prefs.ArticleFontSize != 0 ||
		prefs.ArticleLineHeight != DefaultArticleLineHeight {
		t.Errorf("default article metrics = %d / %v, want 0 / %v",
			prefs.ArticleFontSize, prefs.ArticleLineHeight, DefaultArticleLineHeight)
	}

	// Zero size means "follow the interface"; valid custom values pass.
	prefs := Prefs{ArticleFontSize: MinArticleFontSize, ArticleLineHeight: MinArticleLineHeight}
	prefs.Normalize()
	if prefs.ArticleFontSize != MinArticleFontSize || prefs.ArticleLineHeight != MinArticleLineHeight {
		t.Errorf("minimum metrics = %d / %v", prefs.ArticleFontSize, prefs.ArticleLineHeight)
	}
	prefs = Prefs{ArticleFontSize: MaxArticleFontSize, ArticleLineHeight: MaxArticleLineHeight}
	prefs.Normalize()
	if prefs.ArticleFontSize != MaxArticleFontSize || prefs.ArticleLineHeight != MaxArticleLineHeight {
		t.Errorf("maximum metrics = %d / %v", prefs.ArticleFontSize, prefs.ArticleLineHeight)
	}

	// An out-of-range line height (including zero from older databases)
	// resets to the default; an out-of-range size falls back to following
	// the interface.
	prefs = Prefs{ArticleFontSize: MinArticleFontSize - 1, ArticleLineHeight: MaxArticleLineHeight + 0.1}
	prefs.Normalize()
	if prefs.ArticleFontSize != 0 || prefs.ArticleLineHeight != DefaultArticleLineHeight {
		t.Errorf("out-of-range metrics = %d / %v, want 0 / %v",
			prefs.ArticleFontSize, prefs.ArticleLineHeight, DefaultArticleLineHeight)
	}

	prefs = Prefs{ArticleLineHeight: math.NaN()}
	prefs.Normalize()
	if prefs.ArticleLineHeight != DefaultArticleLineHeight {
		t.Errorf("NaN line height = %v, want %v", prefs.ArticleLineHeight, DefaultArticleLineHeight)
	}
}

func TestResolveFontFamily(t *testing.T) {
	if got := ResolveFontFamily(FontFamilySerif); !strings.Contains(got, "Georgia") {
		t.Errorf("serif stack = %q", got)
	}
	if got := ResolveFontFamily(""); !strings.Contains(got, "system-ui") {
		t.Errorf("empty stack = %q", got)
	}
	custom := `"Microsoft YaHei", Arial`
	if got := ResolveFontFamily(custom); got != custom {
		t.Errorf("custom stack = %q, want %q", got, custom)
	}
}
