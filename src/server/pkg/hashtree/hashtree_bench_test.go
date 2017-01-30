// Benchmarks for the hashtree library. How long operations take can depend
// heavily on how much rehashing they do. All times are measured on msteffen's
// Dell XPS laptop with 8 cores and 16GB of RAM.
//
// TODO(msteffen): perhaps repeat experiments on GCP, in case they need to be
// tested later.

package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
)

// BenchmarkClone is idential to BenchmarkDelete, except that it doesn't
// actually delete any data. The idea is to provide a baseline for how long it
// takes to clone a HashTree with 'cnt' elements, so that that number can be
// subtracted from BenchmarkDelete (since that operation is necessarily part of
// the benchmark)
//
// Benchmarked times at rev. 3ecd3d7520b75b0650f69b3cf4d4ea44908255f8
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.003 s/op
// 10k  | 0.042 s/op
// 100k | 0.419 s/op
func BenchmarkClone(b *testing.B) {
	// Add 10k files
	cnt := int(1e4)
	h := HashTree{}
	r := rand.New(rand.NewSource(0))
	for i := 0; i < cnt; i++ {
		h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			br(fmt.Sprintf(`block{hash:"%x"}`, r.Uint32())))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h2 := proto.Clone(&h).(*HashTree)
		h2.DeleteFile("/foo")
	}
}

// Benchmarked times pre-commit (when DeleteFile was recursive, and deleting a
// directory with N files caused N re-hashes up to the root)
//  cnt |  time (s)
// -----+-------------
// 1k   |    0.302 s/op
// 10k  |   37.734 s/op
// 100k | 4390.167 s/op (1.21h)
//
// Benchmarked times at rev. 3ecd3d7520b75b0650f69b3cf4d4ea44908255f8
//  cnt |  time (s)
// -----+-------------
// 1k   | 0.004 s/op
// 10k  | 0.039 s/op
// 100k | 0.476 s/op
func BenchmarkDelete(b *testing.B) {
	// Add 10k files
	cnt := int(1e5)
	h := HashTree{}
	r := rand.New(rand.NewSource(0))
	for i := 0; i < cnt; i++ {
		h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			br(fmt.Sprintf(`block{hash:"%x"}`, r.Uint32())))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h2 := proto.Clone(&h).(*HashTree)
		h2.DeleteFile("/foo")
	}
}

// Benchmarked times at rev. 3ecd3d7520b75b0650f69b3cf4d4ea44908255f8
//  cnt |  time (s)
// -----+-------------
// 1k   |    0.393 s/op
// 10k  |   41.266 s/op
// 100k | 5553.706 s/op (~1.5h)
func BenchmarkMerge(b *testing.B) {
	// Merge 10k trees, each with 1 file (simulating a job)
	cnt := int(1e4)
	trees := make([]Interface, cnt)
	r := rand.New(rand.NewSource(0))
	for i := 0; i < cnt; i++ {
		trees[i] = new(HashTree)
		trees[i].PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			br(fmt.Sprintf(`block{hash:"%x"}`, r.Uint32())))
	}

	h := HashTree{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Merge(trees)
		h = HashTree{}
	}
}
