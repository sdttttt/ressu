//go:build debug

package logging

import "log/slog"

// defaultLevel is the verbosity used when Open receives no explicit level.
const defaultLevel = slog.LevelDebug

// Debug records detailed diagnostics in builds made with the debug tag.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}
