// Package limitreader provides a reader that rejects input larger than a
// configured limit.
package limitreader

import (
	"errors"
	"io"
)

// ErrTooLarge reports that the input exceeded the configured limit.
var ErrTooLarge = errors.New("size limit exceeded")

// Reader reads from an underlying reader while enforcing a byte limit.
type Reader struct {
	reader    io.Reader
	remaining int64
	err       error
}

// New returns a Reader that accepts at most limit bytes from reader. A
// negative limit is treated as zero.
func New(reader io.Reader, limit int64) *Reader {
	if limit < 0 {
		limit = 0
	}
	return &Reader{reader: reader, remaining: limit}
}

// Read reads up to len(p) bytes. It reads at most one byte beyond the limit
// so an oversized input can be rejected instead of silently truncated.
func (r *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.err != nil {
		return 0, r.err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining+1]
	}
	n, err := r.reader.Read(p)
	if int64(n) > r.remaining {
		n = int(r.remaining)
		err = ErrTooLarge
	}
	r.remaining -= int64(n)
	if err != nil && !errors.Is(err, io.EOF) {
		r.err = err
	}
	return n, err
}

// Err returns the first non-EOF error encountered while reading.
func (r *Reader) Err() error {
	return r.err
}
