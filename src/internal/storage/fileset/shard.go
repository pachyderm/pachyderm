package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// Shard shards the file set into path ranges.
// TODO This should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, ids []ID, cb ShardCallback) error {
	fs, err := s.Open(ctx, ids)
	if err != nil {
		return err
	}
	return shard(ctx, fs, s.shardThreshold, cb)
}

// ShardCallback is a callback that returns a path range for each shard.
type ShardCallback func(*index.PathRange) error

// shard creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the content for a path range is greater than
// the passed in shard threshold.
// For each shard, the callback is called with the path range for the shard.
func shard(ctx context.Context, fs FileSet, shardThreshold int64, cb ShardCallback) error {
	var size int64
	pathRange := &index.PathRange{}
	if err := fs.Iterate(ctx, func(f File) error {
		// A shard is created when we have encountered more than shardThreshold content bytes.
		if size >= shardThreshold {
			pathRange.Upper = f.Index().Path
			if err := cb(pathRange); err != nil {
				return err
			}
			size = 0
			pathRange = &index.PathRange{
				Lower: f.Index().Path,
			}
		}
		size += index.SizeBytes(f.Index())
		return nil
	}); err != nil {
		return err
	}
	return cb(pathRange)
}
