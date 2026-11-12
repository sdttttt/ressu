// Package logging owns the local diagnostic log used by the Go side.
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// FileName is the active log file in the user data directory.
	FileName = "ressu.log"

	// maxBackupGenerations is the number of rotated logs kept in addition to
	// the active file, for five logs in total.
	maxBackupGenerations = 4

	// rotatedFileName stores the previous active log.
	rotatedFileName = FileName + ".1"

	defaultMaxBytes = 1 << 20
)

// Options controls the local log file.
type Options struct {
	// MaxBytes rotates the active file once it reaches this size. Zero uses
	// the 1 MiB default; a negative value disables rotation.
	MaxBytes int64

	// Level filters entries before they reach the file. Nil uses Info.
	Level slog.Leveler
}

// File owns the append-only log sink. Its Logger is safe for concurrent use.
type File struct {
	file   *os.File
	logger *slog.Logger
	path   string
}

// Open creates the log directory and appends to its local log file. A file at
// MaxBytes is rotated, keeping the active log and four backups.
func Open(dir string, options Options) (*File, error) {
	if options.MaxBytes == 0 {
		options.MaxBytes = defaultMaxBytes
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("could not create the log directory: %w", err)
	}

	path := filepath.Join(dir, FileName)
	if err := pruneBackups(path); err != nil {
		return nil, err
	}
	if err := rotate(path, options.MaxBytes); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("could not open the log file: %w", err)
	}
	level := options.Level
	if level == nil {
		level = defaultLevel
	}
	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level}))
	return &File{file: file, logger: logger, path: path}, nil
}

// Logger returns the structured logger backed by the log file.
func (f *File) Logger() *slog.Logger {
	return f.logger
}

// Path returns the active log file path.
func (f *File) Path() string {
	return f.path
}

// Close flushes and closes the log file.
func (f *File) Close() error {
	if f == nil || f.file == nil {
		return nil
	}
	err := f.file.Sync()
	if closeErr := f.file.Close(); err == nil {
		err = closeErr
	}
	f.file = nil
	return err
}

// rotate moves an oversized active log to the oldest backup slot, shifting
// existing backups toward it.
func rotate(path string, maxBytes int64) error {
	if maxBytes < 0 {
		return nil
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not inspect the log file: %w", err)
	}
	if info.Size() < maxBytes {
		return nil
	}

	oldest := fmt.Sprintf("%s.%d", path, maxBackupGenerations)
	if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not replace the old log: %w", err)
	}
	for generation := maxBackupGenerations - 1; generation >= 1; generation-- {
		from := fmt.Sprintf("%s.%d", path, generation)
		to := fmt.Sprintf("%s.%d", path, generation+1)
		if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not shift the old logs: %w", err)
		}
	}
	if err := os.Rename(path, path+".1"); err != nil {
		return fmt.Errorf("could not rotate the log file: %w", err)
	}
	return nil
}

// pruneBackups deletes the oldest rotated logs beyond the five-log limit.
func pruneBackups(path string) error {
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("could not inspect the old logs: %w", err)
	}

	generations := make([]int, 0)
	for _, entry := range entries {
		suffix, found := strings.CutPrefix(entry.Name(), FileName+".")
		if !found {
			continue
		}
		generation, err := strconv.Atoi(suffix)
		if err != nil || generation <= 0 {
			continue
		}
		generations = append(generations, generation)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(generations)))

	extra := len(generations) - maxBackupGenerations
	for index := 0; index < extra; index++ {
		generation := generations[index]
		old := fmt.Sprintf("%s.%d", path, generation)
		if err := os.Remove(old); err != nil {
			return fmt.Errorf("could not delete the old log: %w", err)
		}
	}
	return nil
}
