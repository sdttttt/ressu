package app

import (
	"os"
	"testing"

	"github.com/sdttttt/ressu/core/internal/safedial"
)

// TestMain lets the integration tests reach their local httptest servers; the
// production clients refuse private-network addresses.
func TestMain(m *testing.M) {
	safedial.AllowPrivate.Store(true)
	os.Exit(m.Run())
}
