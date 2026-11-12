//go:build !debug

package logging

import (
	"os"
	"testing"
)

func TestDebugIsNoOpWithoutBuildTag(t *testing.T) {
	file, err := Open(t.TempDir(), Options{MaxBytes: -1})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	Debug("debug diagnostic", "operation", "media.test")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("debug build tag unexpectedly logged: %s", data)
	}
}
