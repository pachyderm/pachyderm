package fileset

import (
	"context"
	"fmt"
	"math"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// IsCompacted returns true if the file sets are already in compacted form.
func (s *Storage) IsCompacted(ctx context.Context, ids []ID) (bool, error) {
	prims, err := s.flattenPrimitives(ctx, ids)
	if err != nil {
		return false, err
	}
	return isCompacted(s.compactionConfig, prims), nil
}

func isCompacted(config *CompactionConfig, prims []*Primitive) bool {
	return config.FixedDelay > int64(len(prims)) || indexOfCompacted(config.LevelFactor, prims) == len(prims)
}

// indexOfCompacted returns the last value of i for which the "compacted relationship" is maintained for all layers[:i+1].
// The "compacted relationship" is defined as leftScore >= (rightScore * factor).
// If there is an element at i+1, it will be the first element which does not satisfy the compacted relationship with i.
func indexOfCompacted(factor int64, prims []*Primitive) int {
	for i := 0; i < len(prims)-1; i++ {
		leftScore := compactionScore(prims[i])
		rightScore := compactionScore(prims[i+1])
		if leftScore < rightScore*factor {
			var score int64
			for j := i; j < len(prims); j++ {
				score += compactionScore(prims[j])
			}
			for ; i > 0; i-- {
				if compactionScore(prims[i-1]) >= score*factor {
					return i
				}
				score += compactionScore(prims[i-1])
			}
			return i
		}
	}
	return len(prims)
}

// compactionScore computes a score for a primitive file set that can be used for making decisions about compaction.
// A higher score means that more work would be involved when including the associated primitive file set in a compaction.
func compactionScore(prim *Primitive) int64 {
	// TODO: Add to prevent full compaction when migrating?
	//	if prim.SizeBytes > 0 {
	//		return prim.SizeBytes
	//	}
	var score int64
	if prim.Additive != nil {
		score += prim.Additive.NumFiles * units.KB
		score += prim.Additive.SizeBytes
	}
	if prim.Deletive != nil {
		score += prim.Deletive.NumFiles * units.KB
	}
	return score
}

// Compact compacts the contents of ids into a new file set with the specified ttl and returns the ID.
// Compact always returns the ID of a primitive file set.
// Compact does not renew ids.
// It is the responsibility of the caller to renew ids.  In some cases they may be permanent and not require renewal.
func (s *Storage) Compact(ctx context.Context, ids []ID, ttl time.Duration, opts ...index.Option) (*ID, error) {
	w := s.newWriter(ctx, WithTTL(ttl))
	fs, err := s.Open(ctx, ids)
	if err != nil {
		return nil, err
	}
	// TODO: Consider adding prefetching here.
	if err := CopyDeletedFiles(ctx, w, fs, opts...); err != nil {
		return nil, err
	}
	if err := CopyFiles(ctx, w, fs, opts...); err != nil {
		return nil, err
	}
	return w.Close()
}

// CompactCallback is the standard callback signature for a compaction operation.
type CompactCallback func(context.Context, *log.Entry, []ID, time.Duration) (*ID, error)

// CompactLevelBased performs a level-based compaction on the passed in file sets.
func (s *Storage) CompactLevelBased(ctx context.Context, logger *log.Entry, ids []ID, maxFanIn int, ttl time.Duration, compact CompactCallback) (*ID, error) {
	ids, err := s.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	prims, err := s.getPrimitives(ctx, ids)
	if err != nil {
		return nil, err
	}
	if isCompacted(s.compactionConfig, prims) {
		return s.Compose(ctx, ids, ttl)
	}
	var id *ID
	if err := s.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *Renewer) error {
		i := indexOfCompacted(s.compactionConfig.LevelFactor, prims)
		if err := miscutil.LogStep(ctx, logger, fmt.Sprintf("compacting %v levels out of %v", len(ids)-i, len(ids)), func() error {
			id, err = s.compactLevels(ctx, logger, ids[i:], maxFanIn, ttl, compact)
			if err != nil {
				return err
			}
			return renewer.Add(ctx, *id)
		}); err != nil {
			return err
		}
		id, err = s.CompactLevelBased(ctx, logger, append(ids[:i], *id), maxFanIn, ttl, compact)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

// compactLevels compacts a list of levels.
// The compaction happens in steps where each step includes file sets of similar compaction score.
// For each step, ranges of file sets in the list that are of length maxFanIn and contain file sets that are
// less than or equal to the step score are compacted.
// For each step, this process is repeated until there are no more ranges eligible for compaction in the step.
// The file sets being compacted must be contiguous because the file operation order matters.
// This algorithm ensures that we compact file sets with lower scores together first before compacting higher score file sets.
func (s *Storage) compactLevels(ctx context.Context, logger *log.Entry, ids []ID, maxFanIn int, ttl time.Duration, compact CompactCallback) (*ID, error) {
	var id *ID
	if err := s.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *Renewer) error {
		for step := 0; len(ids) > maxFanIn; step++ {
			if err := miscutil.LogStep(ctx, logger, fmt.Sprintf("processing step %v", step), func() error {
				stepScore := s.stepScore(step)
				var emptyStep bool
				for !emptyStep {
					emptyStep = true
					nextIds := make([]ID, 0, len(ids))
					var compactIds []ID
					eg, ctx := errgroup.WithContext(ctx)
					for _, id := range ids {
						prim, err := s.getPrimitive(ctx, id)
						if err != nil {
							return err
						}
						compactIds = append(compactIds, id)
						if compactionScore(prim) > stepScore {
							nextIds = append(nextIds, compactIds...)
							compactIds = nil
							continue
						}
						if len(compactIds) == maxFanIn {
							emptyStep = false
							ids := compactIds
							i := len(nextIds)
							nextIds = append(nextIds, ID{})
							eg.Go(func() error {
								logger := logger.WithFields(log.Fields{
									"batch": uuid.NewWithoutDashes(),
								})
								return miscutil.LogStep(ctx, logger, "compacting batch", func() error {
									id, err := compact(ctx, logger, ids, ttl)
									if err != nil {
										return err
									}
									if err := renewer.Add(ctx, *id); err != nil {
										return err
									}
									nextIds[i] = *id
									return nil
								})
							})
							compactIds = nil
						}
					}
					nextIds = append(nextIds, compactIds...)
					if err := eg.Wait(); err != nil {
						return errors.EnsureStack(err)
					}
					ids = nextIds
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if len(ids) == 1 {
			id = &ids[0]
			return nil
		}
		logger = logger.WithFields(log.Fields{
			"batch": uuid.NewWithoutDashes(),
		})
		var err error
		id, err = compact(ctx, logger, ids, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (s *Storage) stepScore(step int) int64 {
	return s.shardSizeThreshold * int64(math.Pow(float64(s.compactionConfig.LevelFactor), float64(step)))
}
