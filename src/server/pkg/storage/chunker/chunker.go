package chunker

import (
	"bytes"
	"io"

	"github.com/chmduquesne/rollinghash/buzhash64"
)

const (
	// MB is Megabytes.
	MB = 1024 * 1024
	// MinSize is the minimum chunk size.
	MinSize = 5 * MB
	// AverageBits determines the average chunk size (2^AverageBits).
	AverageBits = 23
	// MaxSize is the maximum chunk size.
	MaxSize = 15 * MB
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

// Chunker is a content defined chunker.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
// A min/max chunk size is also enforced to avoid creating chunks that are too small/big.
type Chunker struct {
	f         func(io.Reader) error
	buf       *bytes.Buffer
	hash      *buzhash64.Buzhash64
	splitMask uint64
}

// New creates a new content defined chunker.
func New(f func(io.Reader) error) *Chunker {
	// Initialize buzhash64 with WindowSize window.
	hash := buzhash64.New()
	hash.Write(make([]byte, WindowSize))
	return &Chunker{
		f:         f,
		buf:       &bytes.Buffer{},
		hash:      hash,
		splitMask: (1 << uint64(AverageBits)) - 1,
	}
}

// Write rolls through the data written, calling c.f when a chunk is found.
// Note: If making changes to this function, be wary of the performance
// implications (check before and after performance with chunker benchmarks).
func (c *Chunker) Write(data []byte) error {
	offset := 0
	size := c.buf.Len()
	for i, b := range data {
		size++
		c.hash.Roll(b)
		// If the AverageBits lowest bits are zero and the chunk is >= MinSize, then a split point is created.
		// If the chunk is >= MaxSize, then a split point is created regardless.
		if c.hash.Sum64()&c.splitMask == 0 && size >= MinSize || size >= MaxSize {
			c.buf.Write(data[offset : i-offset+1])
			if err := c.f(c.buf); err != nil {
				return err
			}
			c.buf.Reset()
			offset = i + 1
			size = 0
		}
	}
	c.buf.Write(data[offset:])
	return nil
}
