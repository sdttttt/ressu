package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/store"
)

// Settings is the preferences service the frontend calls.
type Settings struct {
	store *store.Store
}

// NewSettings serves the preferences of st.
func NewSettings(st *store.Store) *Settings {
	return &Settings{store: st}
}

// Get returns the preferences.
func (s *Settings) Get() store.Prefs {
	return s.store.Prefs()
}

// Update stores the preferences and applies the appearance they ask for.
func (s *Settings) Update(prefs store.Prefs) (store.Prefs, error) {
	if err := s.store.SetPrefs(prefs); err != nil {
		slog.Warn("preference update failed", "operation", "settings.update", "error", err)
		return store.Prefs{}, err
	}
	ApplyTheme(prefs.Theme)
	return s.store.Prefs(), nil
}

// DataDir returns the directory holding the app's data.
func (s *Settings) DataDir() string {
	return s.store.Dir()
}

// OpenDataDir shows the data directory in the file manager.
func (s *Settings) OpenDataDir() error {
	if err := mygo.Shell.OpenPath(s.store.Dir()); err != nil {
		slog.Warn("data directory open failed", "operation", "settings.openDataDir", "error", err)
		return err
	}
	return nil
}

// OpenLink opens a post's link in the browser, or in a window of the app when
// the preferences ask for that.
func (s *Settings) OpenLink(ctx context.Context, link string) error {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil
	}
	if s.store.Prefs().OpenLinks == store.OpenLinksWindow {
		mygo.NewWindow(mygo.WindowOptions{Title: link, URL: link, Width: 1000, Height: 800})
		return nil
	}
	if err := mygo.Shell.OpenExternal(link); err != nil {
		slog.Warn("external link open failed",
			"operation", "settings.openLink", "externalHost", feedHost(link), "error", err)
		return err
	}
	return nil
}

// ApplyTheme makes the app and its windows follow theme, one of "system",
// "light", "dark" or "sepia". Sepia counts as the dark appearance, so the
// native chrome follows dark; the page itself keeps its light palette and
// only paints a warm, eye-friendly background.
func ApplyTheme(theme string) {
	switch theme {
	case store.ThemeLight:
		mygo.Theme.SetSource(mygo.ThemeLight)
	case store.ThemeDark, store.ThemeSepia:
		mygo.Theme.SetSource(mygo.ThemeDark)
	default:
		mygo.Theme.SetSource(mygo.ThemeSystem)
	}
}
