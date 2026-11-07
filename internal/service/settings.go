package service

import (
	"context"
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
	return mygo.Shell.OpenPath(s.store.Dir())
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
	return mygo.Shell.OpenExternal(link)
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
