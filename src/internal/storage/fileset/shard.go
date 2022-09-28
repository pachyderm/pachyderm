package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// Shard shards the file set into path ranges.
// TODO This should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, ids []ID, pathRange *index.PathRange) ([]*index.PathRange, error) {
	fs, err := s.Open(ctx, ids)
	if err != nil {
		return nil, err
	}
	return shard(ctx, fs, s.shardCountThreshold, s.shardSizeThreshold, pathRange)
}

// shard creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the files is greater than or equal to the passed
// in size threshold or the file count is greater than or equal to the passed in count threshold.
func shard(ctx context.Context, fs FileSet, countThreshold, sizeThreshold int64, basePathRange *index.PathRange) ([]*index.PathRange, error) {
	var shards []*index.PathRange
	pathRange := &index.PathRange{
		Lower: basePathRange.Lower,
	}
	var prev string
	var count, size int64
	if err := fs.Iterate(ctx, func(f File) error {
		if prev != f.Index().Path && (count >= countThreshold || size >= sizeThreshold) {
			pathRange.Upper = f.Index().Path
			shards = append(shards, pathRange)
			pathRange = &index.PathRange{
				Lower: f.Index().Path,
			}
			count = 0
			size = 0
		}
		prev = f.Index().Path
		count++
		size += index.SizeBytes(f.Index())
		return nil
	}, index.WithRange(basePathRange)); err != nil {
		return nil, errors.EnsureStack(err)
	}
	pathRange.Upper = basePathRange.Upper
	shards = append(shards, pathRange)
	return shards, nil
}
