// Benchmarks for the hashtree library. How long operations take can depend
// heavily on how much rehashing they do. All times are measured on msteffen's
// Dell XPS laptop with 8 cores and 16GB of RAM.
//
// TODO(msteffen): repeat experiments on GCP, in case they need to be
// reproduced later (though times shouldn't vary all that much, modulo a few
// milliseconds. The important thing is to limit # hashes for large trees,
// which can cause these tests to go from <5s to >1h if done wrong).

package hashtree

import (
	"fmt"
	"math/rand"
	"testing"
)

// BenchmarkPutFile tests the amount of time it takes to PutFile 'cnt' files
// into a HashTree.
//
// Because "/foo" and "/" in 'h' must be rehashed after each PutFile, and the
// amount of time it takes to do the rehashing is proportional to the number of
// files already in 'h', this is an O(n^2) operation with respect to 'cnt'.
// Because of this, BenchmarkPutFile can be very slow for large 'cnt', often
// much slower than BenchmarkMerge. Be sure to set -timeout 3h for 'cnt' == 100k
//
// Benchmarked times at rev. 6b8e9df38e42f624d2da0aaa785753e9e1d68c0d
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.006 s/op
// 10k  | 0.068 s/op
// 100k | 0.813 s/op
func benchmarkPutFileN(b *testing.B, cnt int) {
	// Add 'cnt' files
	r := rand.New(rand.NewSource(0))
	for n := 0; n < b.N; n++ {
		h := newHashTree(b)
		for i := 0; i < cnt; i++ {
			h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
				obj(fmt.Sprintf(`hash:"%x"`, r.Uint32())), 1)
		}
		h.Hash()
	}
}

func BenchmarkPutFile1k(b *testing.B) {
	benchmarkPutFileN(b, 1e3)
}

func BenchmarkPutFile10k(b *testing.B) {
	benchmarkPutFileN(b, 1e4)
}

func BenchmarkPutFile100k(b *testing.B) {
	benchmarkPutFileN(b, 1e5)
}

// BenchmarkMerge measures how long it takes to merge 'cnt' trees, each of which
// has a single small file, into one central hash tree. This is similar to what
// happens at the completion of a job. Because all re-hashing is saved until the
// end, this is O(n) with respect to 'cnt', making it much faster than calling
// PutFile 'cnt' times.
//
// Benchmarked times at rev. 6b8e9df38e42f624d2da0aaa785753e9e1d68c0d
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.006 s/op
// 10k  | 0.082 s/op
// 100k | 2.750 s/op
func benchmarkMergeN(b *testing.B, cnt int) {
	// Merge 'cnt' trees, each with 1 file (simulating a job)
	trees := make([]HashTree, cnt)
	r := rand.New(rand.NewSource(0))
	var err error
	for i := 0; i < cnt; i++ {
		t := newHashTree(b)
		t.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			obj(fmt.Sprintf(`hash:"%x"`, r.Uint32())), 1)
		t.Hash()
		if err != nil {
			b.Fatal("could not run benchmark: " + err.Error())
		}
		trees[i] = t
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := newHashTree(b)
		h.Merge(trees...)
		h.Hash()
	}
}

func BenchmarkMerge1k(b *testing.B) {
	benchmarkMergeN(b, 1e3)
}

func BenchmarkMerge10k(b *testing.B) {
	benchmarkMergeN(b, 1e4)
}

func BenchmarkMerge100k(b *testing.B) {
	benchmarkMergeN(b, 1e5)
}

// BenchmarkClone is idential to BenchmarkDelete, except that it doesn't
// actually call DeleteFile. The idea is to provide a baseline for how long it
// takes to clone a HashTree with 'cnt' elements, so that that number can be
// subtracted from BenchmarkDelete (since that operation is necessarily part of
// the benchmark)
//
// Benchmarked times at rev. 6b8e9df38e42f624d2da0aaa785753e9e1d68c0d
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.002 s/op
// 10k  | 0.028 s/op
// 100k | 0.346 s/op
func benchmarkCloneN(b *testing.B, cnt int) {
	// Create a tree with 'cnt' files
	r := rand.New(rand.NewSource(0))
	h := newHashTree(b)
	for i := 0; i < cnt; i++ {
		h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			obj(fmt.Sprintf(`hash:"%x"`, r.Uint32())), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Copy()
	}
}

func BenchmarkClone1k(b *testing.B) {
	benchmarkCloneN(b, 1e3)
}

func BenchmarkClone10k(b *testing.B) {
	benchmarkCloneN(b, 1e4)
}

func BenchmarkClone100k(b *testing.B) {
	benchmarkCloneN(b, 1e5)
}

// BenchmarkDelete measures how long it takes to delete a directory with 'cnt'
// children from a HashTree. If implemented poorly, this can be a quadratic
// operation (have to re-hash /foo after deleting each /foo/shard-xxxxx) and
// will take >1h to delete /foo containing 100k files
//
// Benchmarked times at rev. 6b8e9df38e42f624d2da0aaa785753e9e1d68c0d
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.002 s/op
// 10k  | 0.030 s/op
// 100k | 0.395 s/op
func benchmarkDeleteN(b *testing.B, cnt int) {
	// Create a tree with 'cnt' files
	r := rand.New(rand.NewSource(0))
	h := newHashTree(b)
	for i := 0; i < cnt; i++ {
		h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			obj(fmt.Sprintf(`hash:"%x"`, r.Uint32())), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h2, err := h.Copy()
		if err != nil {
			b.Fatal("could not clone hashtree in BenchmarkDelete")
		}
		h2.DeleteFile("/foo")
	}
}

func BenchmarkDelete1k(b *testing.B) {
	benchmarkDeleteN(b, 1e3)
}

func BenchmarkDelete10k(b *testing.B) {
	benchmarkDeleteN(b, 1e4)
}

func BenchmarkDelete100k(b *testing.B) {
	benchmarkDeleteN(b, 1e5)
}
