package chunk

import (
	"math"
	"os"
	"path/filepath"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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

// WithObjectCache adds a cache around the currently configured object client
func WithObjectCache(fastLayer obj.Client, size int) StorageOption {
	return func(s *Storage) {
		s.objClient = obj.NewCacheClient(s.objClient, fastLayer, size)
	}
}

// WithSecret sets the secret used to generate chunk encryption keys
func WithSecret(secret []byte) StorageOption {
	return func(s *Storage) {
		s.createOpts.Secret = append([]byte{}, secret...)
	}
}

// WithCompression sets the compression algorithm used to compress chunks
func WithCompression(algo CompressionAlgo) StorageOption {
	return func(s *Storage) {
		s.createOpts.Compression = algo
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

// StorageOptions returns the chunk storage options for the config.
func StorageOptions(conf *serviceenv.Configuration) ([]StorageOption, error) {
	var opts []StorageOption
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
	if conf.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes()))
		if err != nil {
			return nil, err
		}
		diskCache = obj.TracingObjClient("DiskCache", diskCache)
		opts = append(opts, WithObjectCache(diskCache, conf.StorageDiskCacheSize))
	}
	return opts, nil
}
