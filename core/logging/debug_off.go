//go:build !debug

package logging

import "log/slog"

// defaultLevel is the verbosity used when Open receives no explicit level.
const defaultLevel = slog.LevelInfo

// Debug does nothing unless the debug build tag is enabled.
func Debug(msg string, args ...any) {}
