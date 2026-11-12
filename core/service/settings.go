package service

import (
	"log/slog"

	"github.com/sdttttt/ressu/core/media"
	"github.com/sdttttt/ressu/core/store"
)

// Settings is the preferences service the frontend calls.
type Settings struct {
	store *store.Store
	media *media.Cache
}

// NewSettings serves the preferences of st.
func NewSettings(st *store.Store) *Settings {
	return NewSettingsWithMedia(st, nil)
}

// NewSettingsWithMedia applies image cache preferences to the media cache.
func NewSettingsWithMedia(st *store.Store, mediaCache *media.Cache) *Settings {
	return &Settings{store: st, media: mediaCache}
}

// Get returns the preferences.
func (s *Settings) Get() store.Prefs {
	return s.store.Prefs()
}

// Update stores the preferences and applies the image cache limit.
func (s *Settings) Update(prefs store.Prefs) (store.Prefs, error) {
	if err := s.store.SetPrefs(prefs); err != nil {
		slog.Warn("preference update failed", "operation", "settings.update", "error", err)
		return store.Prefs{}, err
	}
	stored := s.store.Prefs()
	if s.media != nil {
		if err := s.media.SetLimit(int64(stored.ImageCacheLimit) << 20); err != nil {
			slog.Warn("image cache limit update failed",
				"operation", "settings.update", "error", err)
			return store.Prefs{}, err
		}
	}
	return stored, nil
}

// StorageInfo describes how much disk space the app occupies.
type StorageInfo struct {
	// DBBytes is the on-disk size of the database file.
	DBBytes int64 `json:"dbBytes"`
	// ImageCacheBytes is the size of the downloaded image cache.
	ImageCacheBytes int64 `json:"imageCacheBytes"`
	// ImageCacheLimitBytes is the active byte limit of the image cache.
	ImageCacheLimitBytes int64 `json:"imageCacheLimitBytes"`
}

// StorageUsage returns the current disk usage of the database and the image
// cache. A missing media cache reports zero image cache bytes.
func (s *Settings) StorageUsage() StorageInfo {
	info := StorageInfo{}
	if size, err := s.store.DBSize(); err == nil {
		info.DBBytes = size
	} else {
		slog.Warn("database size read failed", "operation", "settings.storageUsage", "error", err)
	}
	if s.media != nil {
		info.ImageCacheBytes, info.ImageCacheLimitBytes = s.media.Usage()
	}
	return info
}

// DataDir returns the directory holding the app's data.
func (s *Settings) DataDir() string {
	return s.store.Dir()
}
