package limitreader_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/sdttttt/ressu/core/internal/limitreader"
)

func TestReaderBoundaries(t *testing.T) {
	exact := limitreader.New(strings.NewReader("abc"), 3)
	got, err := io.ReadAll(exact)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abc" {
		t.Errorf("read = %q, want %q", got, "abc")
	}

	over := limitreader.New(strings.NewReader("abcd"), 3)
	if _, err := io.ReadAll(over); !errors.Is(err, limitreader.ErrTooLarge) {
		t.Fatalf("error = %v, want limitreader.ErrTooLarge", err)
	}
}

func TestReaderPreservesReadError(t *testing.T) {
	boom := errors.New("read failed")
	reader := limitreader.New(iotest.ErrReader(boom), 10)
	if _, err := io.ReadAll(reader); !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
	if !errors.Is(reader.Err(), boom) {
		t.Fatalf("stored error = %v, want %v", reader.Err(), boom)
	}
}

func TestReaderTreatsNegativeLimitAsZero(t *testing.T) {
	reader := limitreader.New(strings.NewReader("a"), -1)
	if _, err := io.ReadAll(reader); !errors.Is(err, limitreader.ErrTooLarge) {
		t.Fatalf("error = %v, want limitreader.ErrTooLarge", err)
	}
}
