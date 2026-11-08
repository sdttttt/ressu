package textmetrics

import "testing"

func TestCount(t *testing.T) {
	cases := []struct {
		name   string
		markup string
		want   int
	}{
		{"empty", "", 0},
		{"chinese", "<p>你好，世界。</p>", 4},
		{"english", "<p>Hello, world 42!</p>", 3},
		{"mixed", "<p>中文 words 混合 text</p>", 6},
		{"ignored markup", `<p>正文</p><script>ignored()</script><style>.a{}</style>`, 2},
		{"entities", `<p>Tom &amp; Jerry</p>`, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Count(c.markup); got != c.want {
				t.Errorf("Count(%q) = %d, want %d", c.markup, got, c.want)
			}
		})
	}
}
