package fileset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func createTestFileSet(tb testing.TB, ctx context.Context, s *Storage, numFiles int, fileSize int, random *rand.Rand) FileSet {
	w := s.NewWriter(ctx)
	for i := 0; i < numFiles; i++ {
		data := randutil.Bytes(random, fileSize)
		require.NoError(tb, w.Add(fmt.Sprintf("%08d", i), DefaultFileDatum, bytes.NewReader(data)))
	}
	id, err := w.Close()
	require.NoError(tb, err)
	fs, err := s.Open(ctx, []ID{*id})
	require.NoError(tb, err)
	return fs
}

func TestPrefetcher(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(t)
	seed := int64(1648577872380609229)
	random := rand.New(rand.NewSource(seed))
	tests := []struct {
		name     string
		numFiles int
		fileSize int
	}{
		{"0B", 1000000, 0},
		{"1KB", 100000, units.KB},
		{"100KB", 1000, 100 * units.KB},
		{"10MB", 10, 10 * units.MB},
		{"100MB", 1, 100 * units.MB},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := createTestFileSet(t, ctx, s, test.numFiles, test.fileSize, random)
			fs = NewPrefetcher(s, fs)
			var i int
			require.NoError(t, fs.Iterate(ctx, func(f File) error {
				p, err := strconv.Atoi(f.Index().Path)
				require.NoError(t, err)
				require.Equal(t, i, p)
				i++
				return nil
			}))
			require.Equal(t, test.numFiles, i)
		})
	}
}

func BenchmarkNoPrefetcher(b *testing.B) {
	benchmarkPrefetcher(b, false)
}

func BenchmarkPrefetcher(b *testing.B) {
	benchmarkPrefetcher(b, true)
}

func benchmarkPrefetcher(b *testing.B, prefetch bool) {
	ctx := context.Background()
	s := newTestStorage(b)
	seed := int64(1648577872380609229)
	random := rand.New(rand.NewSource(seed))
	benchmarks := []struct {
		name     string
		numFiles int
		fileSize int
	}{
		{"1KB", 100000, units.KB},
		{"100KB", 1000, 100 * units.KB},
		{"10MB", 10, 10 * units.MB},
		{"100MB", 1, 100 * units.MB},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			fs := createTestFileSet(b, ctx, s, benchmark.numFiles, benchmark.fileSize, random)
			if prefetch {
				fs = NewPrefetcher(s, fs)
			}
			b.SetBytes(int64(benchmark.numFiles * benchmark.fileSize))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				require.NoError(b, fs.Iterate(ctx, func(f File) error {
					return f.Content(ctx, io.Discard)
				}))
			}
		})
	}
}
