package gui

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// formatRelativeTime formats an article time exactly like the webview list.
func formatRelativeTime(value time.Time, now time.Time) string {
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

// formatReadingMinutes estimates reading time from an article word count.
func formatReadingMinutes(count int) string {
	minutes := max(1, int(math.Ceil(float64(count)/250)))
	if minutes < 60 {
		return fmt.Sprintf("约 %d 分钟", minutes)
	}
	return fmt.Sprintf("约 %d 小时", minutes/60)
}

// formatDateTime formats an article timestamp with its absolute local time.
func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.In(time.Local).Format("2006年1月2日 15:04")
}

// formatWordCount formats an article word count.
func formatWordCount(count int) string {
	digits := strconv.Itoa(count)
	for index := len(digits) - 3; index > 0; index -= 3 {
		digits = digits[:index] + "," + digits[index:]
	}
	return digits + " 字"
}
