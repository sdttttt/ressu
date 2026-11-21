package extract_test

import (
	"os"
	"testing"

	"github.com/sdttttt/ressu/core/internal/safedial"
)

func TestMain(m *testing.M) {
	safedial.AllowPrivate.Store(true)
	os.Exit(m.Run())
}
