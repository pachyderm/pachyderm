package fileset

import (
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// StorageOption configures a storage.
type StorageOption func(*Storage)

// WithMemoryThreshold sets the memory threshold that must
// be met before a file set part is serialized (excluding close).
func WithMemoryThreshold(threshold int64) StorageOption {
	return func(s *Storage) {
		s.memThreshold = threshold
	}
}

// WithShardThreshold sets the size threshold that must
// be met before a shard is created by the shard function.
func WithShardThreshold(threshold int64) StorageOption {
	return func(s *Storage) {
		s.shardThreshold = threshold
	}
}

// WithLevelFactor sets the factor which the size of levels in inc
func WithLevelFactor(x int64) StorageOption {
	return func(s *Storage) {
		s.compactionConfig.LevelFactor = x
	}
}

// WithMaxOpenFileSets sets the maximum number of filesets that will be open
// (potentially buffered in memory) at a time.
func WithMaxOpenFileSets(max int) StorageOption {
	return func(s *Storage) {
		s.filesetSem = semaphore.NewWeighted(int64(max))
	}
}

// UnorderedWriterOption configures an UnorderedWriter.
type UnorderedWriterOption func(*UnorderedWriter)

// WithRenewal configures the UnorderedWriter to renew subfileset paths
// with the provided renewer.
func WithRenewal(ttl time.Duration, r *Renewer) UnorderedWriterOption {
	return func(uw *UnorderedWriter) {
		uw.ttl = ttl
		uw.renewer = r
	}
}

// WithParentID sets a factory for the parent fileset ID for the unordered writer.
// This is used for converting directory deletions into a set of point
// deletes for the files contained within the directory.
func WithParentID(getParentID func() (*ID, error)) UnorderedWriterOption {
	return func(uw *UnorderedWriter) {
		uw.getParentID = getParentID
	}
}

// WithValidator sets the validator for paths being written to the unordered writer.
func WithValidator(validator func(string) error) UnorderedWriterOption {
	return func(uw *UnorderedWriter) {
		uw.validator = validator
	}
}

// WithCompact sets the unordered writer to compact the created file sets if they exceed the passed in fan in.
func WithCompact(maxFanIn int) UnorderedWriterOption {
	return func(uw *UnorderedWriter) {
		uw.maxFanIn = maxFanIn
	}
}

// WriterOption configures a file set writer.
type WriterOption func(w *Writer)

// WithIndexCallback sets a function to be called after each index is written.
// If WithNoUpload is set, the function is called after the index would have been written.
func WithIndexCallback(cb func(*index.Index) error) WriterOption {
	return func(w *Writer) {
		w.indexFunc = cb
	}
}

// WithTTL sets the ttl for the fileset
func WithTTL(ttl time.Duration) WriterOption {
	return func(w *Writer) {
		w.ttl = ttl
	}
}

// StorageOptions returns the fileset storage options for the config.
func StorageOptions(conf *serviceenv.StorageConfiguration) []StorageOption {
	var opts []StorageOption
	if conf.StorageMemoryThreshold > 0 {
		opts = append(opts, WithMemoryThreshold(conf.StorageMemoryThreshold))
	}
	if conf.StorageShardThreshold > 0 {
		opts = append(opts, WithShardThreshold(conf.StorageShardThreshold))
	}
	if conf.StorageLevelFactor > 0 {
		opts = append(opts, WithLevelFactor(conf.StorageLevelFactor))
	}
	return opts
}
