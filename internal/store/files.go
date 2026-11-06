package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// readJSON decodes path into v, reporting whether the file exists.
func readJSON[T any](path string) (v T, found bool, err error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return v, false, nil
	}
	if err != nil {
		return v, false, err
	}
	if len(data) == 0 {
		return v, true, nil
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return v, true, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return v, true, nil
}

// writeJSON replaces path with v: the file is written beside it and renamed
// over it, so a crash never leaves half a file behind.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
