// Package buildinfo reads metadata recorded in the running Go build.
package buildinfo

import (
	"os"
	"runtime/debug"
	"time"
)

// Date returns the date associated with the running build. Go records the
// source commit time as vcs.time, so packaged builds use that date. Without
// VCS metadata it falls back to the executable's modification time.
func Date() time.Time {
	if info, ok := debug.ReadBuildInfo(); ok {
		if date := dateFromSettings(info.Settings); !date.IsZero() {
			return date
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return time.Time{}
	}
	file, err := os.Stat(executable)
	if err != nil {
		return time.Time{}
	}
	return file.ModTime().Local()
}

// dateFromSettings returns the source commit time recorded by the Go
// toolchain, or the zero time when the build has no valid vcs.time setting.
func dateFromSettings(settings []debug.BuildSetting) time.Time {
	for _, setting := range settings {
		if setting.Key != "vcs.time" {
			continue
		}
		date, err := time.Parse(time.RFC3339, setting.Value)
		if err != nil {
			continue
		}
		return date.Local()
	}
	return time.Time{}
}
