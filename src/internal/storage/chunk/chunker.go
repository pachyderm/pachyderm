package chunk

import (
	io "io"
	"math"

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
	defaultAverageBits  = 23
	defaultSeed         = 1
	defaultMinChunkSize = 1 * units.MB
	defaultMaxChunkSize = 20 * units.MB
)

type chunkSize struct {
	min, avg, max int
}

// TODO: Expose configuration.
func ComputeChunks(r io.Reader, cb func([]byte) error) error {
	buf := make([]byte, units.MB)
	hash := buzhash64.NewFromUint64Array(buzhash64.GenerateHashes(defaultSeed))
	resetHash(hash)
	splitMask := uint64((1 << uint64(defaultAverageBits)) - 1)
	var chunkBuf []byte
	chunkSize := &chunkSize{
		min: defaultMinChunkSize,
		avg: int(math.Pow(2, float64(defaultAverageBits))),
		max: defaultMaxChunkSize,
	}
	for {
		n, err := r.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		data := buf[:n]
		offset := 0
		for i, b := range data {
			hash.Roll(b)
			if hash.Sum64()&splitMask == 0 {
				if len(chunkBuf)+len(data[offset:i+1]) < chunkSize.min {
					continue
				}
				chunkBuf = append(chunkBuf, data[offset:i+1]...)
				if err := cb(chunkBuf); err != nil {
					return err
				}
				resetHash(hash)
				chunkBuf = nil
				offset = i + 1
				continue
			}
			// TODO: This can be optimized a bit by accounting for it before rolling the data.
			if len(chunkBuf)+len(data[offset:i+1]) >= chunkSize.max {
				chunkBuf = append(chunkBuf, data[offset:i+1]...)
				if err := cb(chunkBuf); err != nil {
					return err
				}
				resetHash(hash)
				chunkBuf = nil
				offset = i + 1
			}
		}
		chunkBuf = append(chunkBuf, data[offset:]...)
		if errors.Is(err, io.EOF) {
			return cb(chunkBuf)
		}
	}
}

func resetHash(hash *buzhash64.Buzhash64) {
	hash.Reset()
	hash.Write(initialWindow)
}
