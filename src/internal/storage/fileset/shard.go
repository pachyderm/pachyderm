package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// Shard shards the file set into path ranges.
// TODO This should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, ids []ID, cb ShardCallback) error {
	fs, err := s.Open(ctx, ids)
	if err != nil {
		return err
	}
	return shard(ctx, fs, s.shardSizeThreshold, s.shardCountThreshold, cb)
}

// ShardCallback is a callback that returns a path range for each shard.
type ShardCallback func(*index.PathRange) error

// shard creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the files is greater than or equal to the passed
// in size threshold or the file count is greater than or equal to the passed in count threshold.
// For each shard, the callback is called with the path range for the shard.
func shard(ctx context.Context, fs FileSet, sizeThreshold, countThreshold int64, cb ShardCallback) error {
	var size, count int64
	pathRange := &index.PathRange{}
	if err := fs.Iterate(ctx, func(f File) error {
		// A shard is created when we have encountered more than shardThreshold content bytes.
		if size >= sizeThreshold || count >= countThreshold {
			if err := cb(pathRange); err != nil {
				return err
			}
			// set new lower bound to include everything after previous range
			pathRange = &index.PathRange{
				Lower:      pathRange.Upper,
				LowerDatum: pathRange.UpperDatum + "_",
			}
			size = 0
			count = 0
		}
		pathRange.Upper = f.Index().Path
		pathRange.UpperDatum = f.Index().File.Datum
		size += index.SizeBytes(f.Index())
		count++
		return nil
	}); err != nil {
		return err
	}
	pathRange.Upper = ""
	return cb(pathRange)
}
