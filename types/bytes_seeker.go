package types

import (
	"fmt"
	"io"
)

// BytesSeeker implements the FileReader interface directly with byte slices
type BytesSeeker struct {
	data []byte
	pos  int64
}

func (b *BytesSeeker) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

// NewBytesSeeker creates a new seeker from a byte slice
func NewBytesSeeker(data []byte) FileReader {
	// Make a copy to avoid external modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return &BytesSeeker{
		data: dataCopy,
		pos:  0,
	}
}

// Read implements io.Reader
func (b *BytesSeeker) Read(p []byte) (n int, err error) {
	if b.pos >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += int64(n)
	return n, nil
}

// Seek implements io.Seeker
func (b *BytesSeeker) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = b.pos + offset
	case io.SeekEnd:
		newPos = int64(len(b.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position: %d", newPos)
	}
	if newPos > int64(len(b.data)) {
		newPos = int64(len(b.data))
	}

	b.pos = newPos
	return newPos, nil
}

// Close implements io.Closer (no-op for bytes)
func (b *BytesSeeker) Close() error {
	return nil
}
