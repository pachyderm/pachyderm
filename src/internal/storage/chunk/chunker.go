package chunk

import (
	"io"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

// initialWindow is the set of bytes used to initialize the window
// of the rolling hash function.
var initialWindow = make([]byte, WindowSize)

const (
	DefaultAverageBits  = 23
	DefaultSeed         = 1
	DefaultMinChunkSize = 1 * units.MB
	DefaultMaxChunkSize = 20 * units.MB
)

// ComputeChunks splits a stream of bytes into chunks using a content-defined
// chunking algorithm. To prevent suboptimal chunk sizes, a minimum and maximum
// chunk size is enforced. This algorithm is useful for ensuring that typical
// data modifications (insertions, deletions, updates) only affect a small
// number of chunks.
// TODO: Expose configuration.
func ComputeChunks(r io.Reader, cb func([]byte) error) error {
	buf := make([]byte, units.MB)
	var chunkBuf []byte
	hash := buzhash64.NewFromUint64Array(buzhash64.GenerateHashes(DefaultSeed))
	resetHash(hash)
	splitMask := uint64((1 << uint64(DefaultAverageBits)) - 1)
	for {
		n, err := r.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return errors.EnsureStack(err)
		}
		data := buf[:n]
		for _, b := range data {
			chunkBuf = append(chunkBuf, b)
			if len(chunkBuf) >= DefaultMaxChunkSize {
				if err := cb(chunkBuf); err != nil {
					return err
				}
				resetHash(hash)
				chunkBuf = nil
				continue
			}
			hash.Roll(b)
			if hash.Sum64()&splitMask == 0 {
				if len(chunkBuf) < DefaultMinChunkSize {
					continue
				}
				if err := cb(chunkBuf); err != nil {
					return err
				}
				resetHash(hash)
				chunkBuf = nil
			}
		}
		if errors.Is(err, io.EOF) {
			return cb(chunkBuf)
		}
	}
}

func resetHash(hash *buzhash64.Buzhash64) {
	hash.Reset()
	hash.Write(initialWindow) //nolint:errcheck
}
