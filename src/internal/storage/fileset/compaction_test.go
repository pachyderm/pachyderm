package fileset

import (
	"context"
	"io"
	"math"
	"math/rand"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	log "github.com/sirupsen/logrus"
)

func TestCompactLevelBasedFuzz(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(t)
	s.shardSizeThreshold = units.KB
	numFileSets := 1000
	maxFanIn := 10
	numSteps := 3
	var stepBases []int64
	for i := 0; i < numSteps; i++ {
		stepBases = append(stepBases, int64(math.Pow(float64(units.KB), float64(i))))
	}
	// Create file sets in runs of maxFanIn to ensure compactions occur at each step.
	var totalScore int64
	var ids []ID
	for i := 0; i < numFileSets/maxFanIn; i++ {
		stepBase := stepBases[rand.Intn(len(stepBases))]
		for j := 0; j < maxFanIn; j++ {
			prim := &Primitive{
				Additive: &index.Index{
					NumFiles:  stepBase,
					SizeBytes: stepBase * units.KB,
				},
				Deletive: &index.Index{
					NumFiles: stepBase,
				},
			}
			totalScore += compactionScore(prim)
			id, err := s.newPrimitive(ctx, prim, track.NoTTL)
			require.NoError(t, err)
			ids = append(ids, *id)
		}
	}
	// Simulate compaction by accumulating the number of files and byte size.
	var count int
	compact := func(ctx context.Context, _ *log.Entry, ids []ID, ttl time.Duration) (*ID, error) {
		if len(ids) > maxFanIn {
			return nil, errors.Errorf("number of file sets being compacted (%v) greater than max fan-in (%v)", len(ids), maxFanIn)
		}
		count++
		additive, deletive := &index.Index{}, &index.Index{}
		for _, id := range ids {
			prim, err := s.getPrimitive(ctx, id)
			if err != nil {
				return nil, err
			}
			additive.NumFiles += prim.Additive.NumFiles
			additive.SizeBytes += prim.Additive.SizeBytes
			deletive.NumFiles += prim.Deletive.NumFiles
		}
		return s.newPrimitive(ctx, &Primitive{
			Additive: additive,
			Deletive: deletive,
		}, ttl)
	}
	logger := log.StandardLogger()
	logger.SetOutput(io.Discard)
	id, err := s.CompactLevelBased(ctx, log.NewEntry(logger), ids, maxFanIn, track.NoTTL, compact)
	require.NoError(t, err)
	// Check the compaction score of the final file set to ensure no file sets were lost.
	prims, err := s.flattenPrimitives(ctx, []ID{*id})
	require.NoError(t, err)
	var checkScore int64
	for _, prim := range prims {
		checkScore += compactionScore(prim)
	}
	require.Equal(t, totalScore, checkScore)
	// Check that compact was called the expected number of times.
	numCompactions := (numFileSets - 1) / (maxFanIn - 1)
	require.Equal(t, numCompactions, count)
}
