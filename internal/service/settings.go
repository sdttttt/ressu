package service

import (
	"log/slog"

	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/store"
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

// DataDir returns the directory holding the app's data.
func (s *Settings) DataDir() string {
	return s.store.Dir()
}
