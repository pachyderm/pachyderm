package fileset

import "github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

// StorageOption configures a storage.
type StorageOption func(s *Storage)

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

//func WithLevelZeroSize(size int64) StorageOption {
//	return func(s *Storage) {
//		s.levelZeroSize = size
//	}
//}
//
//func WitLevelSizeBase(base int) StorageOption {
//	return func(s *Storage) {
//		s.levelSizeBase = base
//	}
//}

// Option configures a file set.
type Option func(f *FileSet)

// WithRoot sets the root path of the file set.
func WithRoot(root string) Option {
	return func(f *FileSet) {
		f.root = root
	}
}

// ServiceEnvToOptions converts a service environment configuration (specifically
// the storage configuration) to a set of storage options.
func ServiceEnvToOptions(env *serviceenv.ServiceEnv) []StorageOption {
	var opts []StorageOption
	if env.StorageMemoryThreshold > 0 {
		opts = append(opts, WithMemoryThreshold(env.StorageMemoryThreshold))
	}
	if env.StorageShardThreshold > 0 {
		opts = append(opts, WithShardThreshold(env.StorageShardThreshold))
	}
	//	if env.StorageLevelZeroSize > 0 {
	//		opts = append(opts, WithLevelZeroSize(env.StorageLevelZeroSize))
	//	}
	//	if env.StorageLevelSizeBase > 0 {
	//		opts = append(opts, WitLevelSizeBase(env.StorageLevelSizeBase))
	//	}
	return opts
}
