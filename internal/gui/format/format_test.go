package format_test

import (
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/gui/format"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Date(2026, time.November, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		value time.Time
		want  string
	}{
		{"zero", time.Time{}, ""},
		{"minutes", now.Add(-5 * time.Minute), "5 分钟前"},
		{"hours", now.Add(-2 * time.Hour), "2 小时前"},
		{"days", now.Add(-3 * 24 * time.Hour), "3 天前"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := format.RelativeTime(test.value, now); got != test.want {
				t.Errorf("RelativeTime() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestFormatArticleMetrics(t *testing.T) {
	if got := format.ReadingMinutes(500); got != "约 2 分钟" {
		t.Errorf("ReadingMinutes() = %q, want %q", got, "约 2 分钟")
	}
	if got := format.WordCount(1234567); got != "1,234,567 字" {
		t.Errorf("WordCount() = %q, want %q", got, "1,234,567 字")
	}
	if got := format.DateTime(time.Time{}); got != "" {
		t.Errorf("DateTime() = %q, want empty", got)
	}
}
