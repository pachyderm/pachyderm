package fileset

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func TestIsCompacted(t *testing.T) {
	ctx := pctx.TestContext(t)
	s := newTestStorage(ctx, t, WithShardSizeThreshold(units.KB))
	prim1 := &Primitive{
		Additive: &index.Index{
			NumFiles:  1,
			SizeBytes: 1 * units.KB,
		},
	}
	prim2 := &Primitive{
		Additive: &index.Index{
			NumFiles:  10,
			SizeBytes: 10 * units.KB,
		},
	}
	prim3 := &Primitive{
		Additive: &index.Index{
			NumFiles:  1000,
			SizeBytes: 1000 * units.KB,
		},
	}
	h1, err := s.newPrimitive(ctx, prim1, track.NoTTL)
	require.NoError(t, err)
	h2, err := s.newPrimitive(ctx, prim2, track.NoTTL)
	require.NoError(t, err)
	h3, err := s.newPrimitive(ctx, prim3, track.NoTTL)
	require.NoError(t, err)
	// Basic failure case
	id, err := s.Compose(ctx, []Handle{*h1, *h2}, track.NoTTL)
	require.NoError(t, err)
	isCompacted, err := s.IsCompacted(ctx, *id)
	require.NoError(t, err)
	require.False(t, isCompacted)
	// Basic success case
	h, err := s.Compose(ctx, []Handle{*h2, *h1}, track.NoTTL)
	require.NoError(t, err)
	isCompacted, err = s.IsCompacted(ctx, *id)
	require.NoError(t, err)
	require.True(t, isCompacted)
	// Success case with composites
	complexId, err := s.Compose(ctx, []Handle{*h3, *h}, track.NoTTL)
	require.NoError(t, err)
	isCompacted, err = s.IsCompacted(ctx, *complexId)
	require.NoError(t, err)
	require.True(t, isCompacted)
}

func TestCompactLevelBasedFuzz(t *testing.T) {
	ctx := pctx.TestContext(t)
	s := newTestStorage(ctx, t, WithShardSizeThreshold(units.KB))
	numFileSets := 1000
	maxFanIn := 10
	numSteps := 3
	var stepBases []int64
	for i := 0; i < numSteps; i++ {
		stepBases = append(stepBases, int64(math.Pow(float64(units.KB), float64(i))))
	}
	// Create file sets in runs of maxFanIn to ensure compactions occur at each step.
	var totalScore int64
	var hs []Handle
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
			h, err := s.newPrimitive(ctx, prim, track.NoTTL)
			require.NoError(t, err)
			hs = append(hs, *h)
		}
	}
	// Simulate compaction by accumulating the number of files and byte size.
	var count int
	compact := func(ctx context.Context, hs []Handle, ttl time.Duration) (*Handle, error) {
		if len(hs) > maxFanIn {
			return nil, errors.Errorf("number of file sets being compacted (%v) greater than max fan-in (%v)", len(hs), maxFanIn)
		}
		count++
		additive, deletive := &index.Index{}, &index.Index{}
		for _, h := range hs {
			prim, err := s.getPrimitive(ctx, h.id)
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
	h, err := s.CompactLevelBased(ctx, hs, maxFanIn, time.Minute, compact)
	require.NoError(t, err)
	// Check the compaction score of the final file set to ensure no file sets were lost.
	prims, err := s.flattenPrimitives(ctx, []Handle{*h})
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

func TestCompactLevelBasedRenewal(t *testing.T) {
	ctx := pctx.TestContext(t)
	s := newTestStorage(ctx, t)
	ttl := 100 * time.Millisecond
	gc := s.NewGC(ttl)
	cancelCtx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(cancelCtx)
	eg.Go(func() error {
		err := gc.RunForever(ctx)
		if errors.Is(context.Cause(cancelCtx), context.Canceled) {
			err = nil
		}
		return err
	})
	numFileSets := 1000
	maxFanIn := 10
	var id *ID
	eg.Go(func() error {
		defer cancel()
		var hs []Handle
		for i := 0; i < numFileSets; i++ {
			prim := &Primitive{
				Additive: &index.Index{
					NumFiles: 1,
				},
			}
			h, err := s.newPrimitive(ctx, prim, track.NoTTL)
			if err != nil {
				return err
			}
			hs = append(hs, *h)
		}
		compact := func(ctx context.Context, hs []Handle, ttl time.Duration) (*Handle, error) {
			time.Sleep(3 * ttl)
			additive := &index.Index{}
			for _, h := range hs {
				prim, err := s.getPrimitive(ctx, h.id)
				if err != nil {
					return nil, err
				}
				additive.NumFiles += prim.Additive.NumFiles
			}
			return s.newPrimitive(ctx, &Primitive{
				Additive: additive,
			}, ttl)
		}
		var err error
		id, err = s.CompactLevelBased(ctx, ids, maxFanIn, ttl, compact)
		return err
	})
	require.NoError(t, eg.Wait())
	ctx = pctx.TestContext(t)
	prims, err := s.flattenPrimitives(ctx, []ID{*id})
	require.NoError(t, err)
	var numFiles int64
	for _, prim := range prims {
		numFiles += prim.Additive.NumFiles
	}
	require.Equal(t, int64(numFileSets), numFiles)
}
