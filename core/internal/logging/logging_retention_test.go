package logging

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenShiftsRotationAndKeepsFiveLogs(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, FileName)
	contents := map[string]string{
		active + ".1": "first",
		active + ".2": "second",
		active + ".3": "third",
		active + ".4": "fourth",
		active:        strings.Repeat("fifth", 10),
	}
	for path, content := range contents {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	file, err := Open(dir, Options{MaxBytes: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.Logger().Warn(strings.Repeat("sixth", 10))
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	expected := map[int]string{
		1: strings.Repeat("fifth", 10),
		2: "first",
		3: "second",
		4: "third",
	}
	for generation, want := range expected {
		path := fmt.Sprintf("%s.%d", active, generation)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("%s = %q, want %q", path, data, want)
		}
	}
	if _, err := os.Stat(active + ".5"); !os.IsNotExist(err) {
		t.Fatalf("%s should not exist", active+".5")
	}
	activeData, err := os.ReadFile(active)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(activeData, []byte("sixth")) {
		t.Fatalf("active log = %q, want sixth", activeData)
	}
}
