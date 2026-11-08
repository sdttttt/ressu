package store

import "testing"

func TestImageCacheLimitBounds(t *testing.T) {
	for _, test := range []struct {
		limit int
		want  int
	}{
		{limit: 0, want: DefaultImageCacheLimit},
		{limit: 31, want: DefaultImageCacheLimit},
		{limit: 32, want: 32},
		{limit: 256, want: 256},
		{limit: 512, want: 512},
		{limit: 513, want: DefaultImageCacheLimit},
	} {
		prefs := Prefs{ImageCacheLimit: test.limit}
		prefs.Normalize()
		if prefs.ImageCacheLimit != test.want {
			t.Fatalf("Normalize(%d) = %d, want %d", test.limit, prefs.ImageCacheLimit, test.want)
		}
	}
}
