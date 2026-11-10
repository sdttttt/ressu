package buildinfo

import (
	"runtime/debug"
	"testing"
	"time"
)

func TestDateFromSettings(t *testing.T) {
	settings := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abc123"},
		{Key: "vcs.time", Value: "2026-11-10T12:34:56+08:00"},
	}

	got := dateFromSettings(settings)
	want, err := time.Parse(time.RFC3339, "2026-11-10T12:34:56+08:00")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("dateFromSettings = %v, want %v", got, want)
	}
}

func TestDateFromSettingsWithoutTimestamp(t *testing.T) {
	settings := []debug.BuildSetting{
		{Key: "vcs.time", Value: "not a timestamp"},
		{Key: "vcs.revision", Value: "abc123"},
	}

	if got := dateFromSettings(settings); !got.IsZero() {
		t.Fatalf("dateFromSettings = %v, want zero time", got)
	}
}
