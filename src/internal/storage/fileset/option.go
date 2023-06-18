package fileset

import (
	"time"

	"golang.org/x/sync/semaphore"
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

// WithShardSizeThreshold sets the size threshold that must
// be met before a shard is created by the shard function.
func WithShardSizeThreshold(threshold int64) StorageOption {
	return func(s *Storage) {
		s.shardConfig.SizeBytes = threshold
	}
}

// WithShardCountThreshold sets the count threshold that must
// be met before a shard is created by the shard function.
func WithShardCountThreshold(threshold int64) StorageOption {
	return func(s *Storage) {
		s.shardConfig.NumFiles = threshold
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

// WithTTL sets the ttl for the fileset
func WithTTL(ttl time.Duration) WriterOption {
	return func(w *Writer) {
		w.ttl = ttl
	}
}
