// Package format renders native UI values in Chinese with webview-compatible
// formatting.
package format

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// RelativeTime formats an article time exactly like the webview list.
func RelativeTime(value time.Time, now time.Time) string {
	if value.IsZero() {
		return ""
	}

	minutes := int(math.Round(now.Sub(value).Minutes()))
	if minutes < 1 {
		return "刚刚"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d 分钟前", minutes)
	}
	hours := int(math.Round(float64(minutes) / 60))
	if hours < 24 {
		return fmt.Sprintf("%d 小时前", hours)
	}
	days := int(math.Round(float64(hours) / 24))
	if days < 30 {
		return fmt.Sprintf("%d 天前", days)
	}
	return value.In(time.Local).Format("2006年1月2日")
}

// ReadingMinutes estimates reading time from an article word count.
func ReadingMinutes(count int) string {
	minutes := max(1, int(math.Ceil(float64(count)/250)))
	if minutes < 60 {
		return fmt.Sprintf("约 %d 分钟", minutes)
	}
	return fmt.Sprintf("约 %d 小时", minutes/60)
}

// DateTime formats an article timestamp with its absolute local time.
func DateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.In(time.Local).Format("2006年1月2日 15:04")
}

// WordCount formats an article word count.
func WordCount(count int) string {
	digits := strconv.Itoa(count)
	for index := len(digits) - 3; index > 0; index -= 3 {
		digits = digits[:index] + "," + digits[index:]
	}
	return digits + " 字"
}

// Bytes formats a byte count with a binary unit, keeping one decimal place
// from megabytes upward.
func Bytes(count int64) string {
	if count < 0 {
		count = 0
	}
	switch {
	case count >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(count)/(1<<30))
	case count >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(count)/(1<<20))
	case count >= 1<<10:
		return fmt.Sprintf("%d KB", count/(1<<10))
	default:
		return fmt.Sprintf("%d B", count)
	}
}
