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
// be met before a shard is created by the distributed merge.
func WithShardThreshold(threshold int64) StorageOption {
	return func(s *Storage) {
		s.shardThreshold = threshold
	}
}

// FileSetOption configures a file set.
type FileSetOption func(f *FileSet)

// WithRoot sets the root path of the file set.
func WithRoot(root string) FileSetOption {
	return func(f *FileSet) {
		f.root = root
	}
}

func ServiceEnvToOptions(env *serviceenv.ServiceEnv) []StorageOption {
	var opts []StorageOption
	if env.StorageMemoryThreshold > 0 {
		opts = append(opts, WithMemoryThreshold(env.StorageMemoryThreshold))
	}
	if env.StorageShardThreshold > 0 {
		opts = append(opts, WithShardThreshold(env.StorageShardThreshold))
	}
	return opts
}
