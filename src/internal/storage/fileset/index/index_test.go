package index

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

func write(tb testing.TB, chunks *chunk.Storage, fileNames []string) *Index {
	iw := NewWriter(context.Background(), chunks, "test")
	for _, fileName := range fileNames {
		idx := &Index{
			Path: fileName,
			File: &File{},
		}
		require.NoError(tb, iw.WriteIndex(idx))
	}
	topIdx, err := iw.Close()
	require.NoError(tb, err)
	return topIdx
}

func actualFiles(tb testing.TB, chunks *chunk.Storage, cache *Cache, topIdx *Index, opts ...Option) []string {
	ir := NewReader(chunks, cache, topIdx, opts...)
	result := []string{}
	require.NoError(tb, ir.Iterate(context.Background(), func(idx *Index) error {
		result = append(result, idx.Path)
		return nil
	}))
	return result
}

func expectedFiles(fileNames []string, prefix string) []string {
	result := []string{}
	for _, fileName := range fileNames {
		if strings.HasPrefix(fileName, prefix) {
			result = append(result, fileName)
		}
	}
	return result
}

func pathRange(fileNames []string) *PathRange {
	return &PathRange{
		Lower: fileNames[0],
		Upper: fileNames[len(fileNames)-1] + "_",
	}
}

func Check(t *testing.T, permString string) {
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	_, chunks := chunk.NewTestStorage(t, db, tr)
	cache := NewCache(chunks, 5)
	fileNames := Generate(permString)
	topIdx := write(t, chunks, fileNames)
	t.Run("Full", func(t *testing.T) {
		expected := fileNames
		actual := actualFiles(t, chunks, cache, topIdx)
		require.Equal(t, expected, actual)
	})
	t.Run("FirstFile", func(t *testing.T) {
		prefix := fileNames[0]
		expected := []string{prefix}
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
	t.Run("FirstRange", func(t *testing.T) {
		prefix := string(fileNames[0][0])
		expected := expectedFiles(fileNames, prefix)
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
	t.Run("MiddleFile", func(t *testing.T) {
		prefix := fileNames[len(fileNames)/2]
		expected := []string{prefix}
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
	t.Run("MiddleRange", func(t *testing.T) {
		prefix := string(fileNames[len(fileNames)/2][0])
		expected := expectedFiles(fileNames, prefix)
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
	t.Run("LastFile", func(t *testing.T) {
		prefix := fileNames[len(fileNames)-1]
		expected := []string{prefix}
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
	t.Run("LastRange", func(t *testing.T) {
		prefix := string(fileNames[len(fileNames)-1][0])
		expected := expectedFiles(fileNames, prefix)
		actual := actualFiles(t, chunks, cache, topIdx, WithPrefix(prefix))
		require.Equal(t, expected, actual)
		actual = actualFiles(t, chunks, cache, topIdx, WithRange(pathRange(expected)))
		require.Equal(t, expected, actual)
	})
}

func TestSingleLevel(t *testing.T) {
	Check(t, "abc")
}

func TestMultiLevel(t *testing.T) {
	Check(t, "abcdefgh")
}

func TestConcatFuzz(t *testing.T) {
	ctx := context.Background()
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	_, chunks := chunk.NewTestStorage(t, db, tr)
	cache := NewCache(chunks, 5)
	// Create ten file sets and concatenate them into one file set.
	var topIdxs []*Index
	var totalNumFiles int
	for i := 0; i < 10; i++ {
		iw := NewWriter(ctx, chunks, uuid.NewWithoutDashes())
		numFiles := rand.Intn(100000)
		for j := 0; j < numFiles; j++ {
			idx := &Index{
				Path:      fmt.Sprintf("%08d", totalNumFiles+j),
				File:      &File{},
				NumFiles:  1,
				SizeBytes: 1,
			}
			require.NoError(t, iw.WriteIndex(idx))
		}
		topIdx, err := iw.Close()
		require.NoError(t, err)
		topIdxs = append(topIdxs, topIdx)
		totalNumFiles += numFiles
	}
	iw := NewWriter(ctx, chunks, uuid.NewWithoutDashes())
	for _, topIdx := range topIdxs {
		require.NoError(t, iw.WriteIndex(topIdx))
	}
	topIdx, err := iw.Close()
	require.NoError(t, err)
	// Check that the correct files exist and that the index metadata is correct.
	ir := NewReader(chunks, cache, topIdx)
	var i int
	require.NoError(t, ir.Iterate(ctx, func(idx *Index) error {
		p, err := strconv.Atoi(idx.Path)
		require.NoError(t, err)
		require.Equal(t, i, p)
		i++
		return nil
	}))
	require.Equal(t, totalNumFiles, i)
	require.Equal(t, int64(totalNumFiles), topIdx.NumFiles)
	require.Equal(t, int64(totalNumFiles), topIdx.SizeBytes)
}

func TestShard(t *testing.T) {
	ctx := context.Background()
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	_, chunks := chunk.NewTestStorage(t, db, tr)
	cache := NewCache(chunks, 5)
	numFiles := 100000
	iw := NewWriter(context.Background(), chunks, uuid.NewWithoutDashes())
	for i := 0; i < numFiles; i++ {
		idx := &Index{
			Path:      fmt.Sprintf("%08d", i),
			File:      &File{},
			NumFiles:  1,
			SizeBytes: 1,
		}
		require.NoError(t, iw.WriteIndex(idx))
	}
	topIdx, err := iw.Close()
	require.NoError(t, err)
	// Check that the correct number of shards are created for a variety of shard configurations.
	check := func(shardConfig *ShardConfig, expectedNumShards int) {
		ir := NewReader(chunks, cache, topIdx, WithShardConfig(shardConfig))
		shards, err := ir.Shards(ctx)
		require.NoError(t, err)
		require.Equal(t, expectedNumShards, len(shards))
	}
	for i := 1; i <= numFiles; i *= 10 {
		check(&ShardConfig{
			NumFiles:  int64(i),
			SizeBytes: math.MaxInt64,
		}, numFiles/i)
		check(&ShardConfig{
			NumFiles:  math.MaxInt64,
			SizeBytes: int64(i),
		}, numFiles/i)
	}
}

func TestShardFuzz(t *testing.T) {
	ctx := context.Background()
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	_, chunks := chunk.NewTestStorage(t, db, tr)
	cache := NewCache(chunks, 5)
	numFiles := 100000
	iw := NewWriter(ctx, chunks, uuid.NewWithoutDashes())
	for i := 0; i < numFiles; i++ {
		idx := &Index{
			Path:      fmt.Sprintf("%08d", i),
			File:      &File{},
			NumFiles:  rand.Int63n(100),
			SizeBytes: rand.Int63n(100),
		}
		require.NoError(t, iw.WriteIndex(idx))
	}
	topIdx, err := iw.Close()
	require.NoError(t, err)
	// Check that the correct files exist for a variety of shard configurations.
	check := func(shardConfig *ShardConfig) {
		ir := NewReader(chunks, cache, topIdx, WithShardConfig(shardConfig))
		shards, err := ir.Shards(ctx)
		require.NoError(t, err)
		var i int
		for _, shard := range shards {
			ir := NewReader(chunks, cache, topIdx, WithRange(shard))
			require.NoError(t, ir.Iterate(ctx, func(idx *Index) error {
				p, err := strconv.Atoi(idx.Path)
				require.NoError(t, err)
				require.Equal(t, i, p)
				i++
				return nil
			}))
		}
	}
	for i := 100; i <= numFiles; i *= 10 {
		check(&ShardConfig{
			NumFiles:  int64(i),
			SizeBytes: int64(i),
		})
	}
}

func BenchmarkMultiLevel(b *testing.B) {
	db := dockertestenv.NewTestDB(b)
	tr := track.NewTestTracker(b, db)
	_, chunks := chunk.NewTestStorage(b, db, tr)
	cache := NewCache(chunks, 5)
	fileNames := Generate("abcdefgh")
	topIdx := write(b, chunks, fileNames)
	for i := 0; i < b.N; i++ {
		for _, fileName := range fileNames {
			ir := NewReader(chunks, cache, topIdx, WithPrefix(fileName))
			require.NoError(b, ir.Iterate(context.Background(), func(idx *Index) error {
				return nil
			}))
		}
	}
}
