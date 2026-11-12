package service

import (
	"testing"

	"github.com/sdttttt/ressu/core/media"
)

func TestStorageUsageReportsDatabaseAndCache(t *testing.T) {
	st := openStore(t)
	cacheDir := t.TempDir() + "/media"
	cache, err := media.Open(cacheDir, media.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	settings := NewSettingsWithMedia(st, cache)
	info := settings.StorageUsage()
	if info.DBBytes <= 0 {
		t.Errorf("db bytes = %d, want a positive size", info.DBBytes)
	}
	if info.ImageCacheBytes != 0 {
		t.Errorf("image cache bytes = %d, want zero for an empty cache", info.ImageCacheBytes)
	}
	if info.ImageCacheLimitBytes != media.DefaultCacheBytes {
		t.Errorf("image cache limit = %d, want default %d", info.ImageCacheLimitBytes, media.DefaultCacheBytes)
	}
}

func TestStorageUsageWithoutMediaCache(t *testing.T) {
	st := openStore(t)

	info := NewSettings(st).StorageUsage()
	if info.DBBytes <= 0 {
		t.Errorf("db bytes = %d, want a positive size", info.DBBytes)
	}
	if info.ImageCacheBytes != 0 || info.ImageCacheLimitBytes != 0 {
		t.Errorf("image cache = %d/%d, want zero without a media cache",
			info.ImageCacheBytes, info.ImageCacheLimitBytes)
	}
}
