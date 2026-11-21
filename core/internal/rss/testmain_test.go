package rss

import (
	"os"
	"testing"

	"github.com/sdttttt/ressu/core/internal/safedial"
)

// TestMain lets the fetch tests reach their local httptest servers; the
// production client refuses private-network addresses.
func TestMain(m *testing.M) {
	safedial.AllowPrivate.Store(true)
	os.Exit(m.Run())
}
