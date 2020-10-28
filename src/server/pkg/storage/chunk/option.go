package chunk

import (
	"math"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// StorageOption configures a storage.
type StorageOption func(s *Storage)

// WithMaxConcurrentObjects sets the maximum number of object writers (upload)
// and readers (download) that can be open at a time.
func WithMaxConcurrentObjects(maxDownload, maxUpload int) StorageOption {
	return func(s *Storage) {
		s.objClient = obj.NewLimitedClient(s.objClient, maxDownload, maxUpload)
	}
}

// WithGCTimeout sets the default chunk ttl for this Storage instance
func WithGCTimeout(timeout time.Duration) StorageOption {
	return func(s *Storage) {
		s.defaultChunkTTL = timeout
	}
}

// WithObjectCache adds a cache around the currently configured object client
func WithObjectCache(fastLayer obj.Client, size int) StorageOption {
	return func(s *Storage) {
		s.objClient = obj.NewCacheClient(s.objClient, fastLayer, size)
	}
}

// WriterOption configures a chunk writer.
type WriterOption func(w *Writer)

// WithRollingHashConfig sets up the rolling hash with the passed in configuration.
func WithRollingHashConfig(averageBits int, seed int64) WriterOption {
	return func(w *Writer) {
		w.chunkSize.avg = int(math.Pow(2, float64(averageBits)))
		w.splitMask = (1 << uint64(averageBits)) - 1
		w.hash = buzhash64.NewFromUint64Array(buzhash64.GenerateHashes(seed))
	}
}

// WithMinMax sets the minimum and maximum chunk size.
func WithMinMax(min, max int) WriterOption {
	return func(w *Writer) {
		w.chunkSize.min = min
		w.chunkSize.max = max
	}
}

// WithNoUpload sets the writer to no upload (will not upload chunks).
func WithNoUpload() WriterOption {
	return func(w *Writer) {
		w.noUpload = true
	}
}
