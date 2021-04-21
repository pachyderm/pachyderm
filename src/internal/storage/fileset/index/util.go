package index

import (
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
)

// Perm calls f with each permutation of a.
func Perm(a []rune, f func([]rune)) {
	perm(a, f, 0)
}

// Permute the values at index i to len(a)-1.
func perm(a []rune, f func([]rune), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}

// Generate generates the permutations of the passed in string and returns them sorted.
func Generate(s string) []string {
	fileNames := []string{}
	Perm([]rune(s), func(fileName []rune) {
		fileNames = append(fileNames, string(fileName))
	})
	sort.Strings(fileNames)
	return fileNames
}

// PointsTo returns a list of all the chunks this index references
func PointsTo(idx *Index) []chunk.ID {
	if idx == nil {
		return nil
	}
	if idx.Range != nil {
		return []chunk.ID{chunk.ID(idx.Range.ChunkRef.Ref.Id)}
	}
	var ids []chunk.ID
	if idx.File != nil {
		for _, dr := range idx.File.DataRefs {
			ids = append(ids, chunk.ID(dr.Ref.Id))
		}
	}
	return ids
}

// SizeBytes computes the size of the indexed data in bytes.
func SizeBytes(idx *Index) int64 {
	var size int64
	if idx == nil {
		return size
	}
	if idx.File != nil {
		for _, dataRef := range idx.File.DataRefs {
			size += dataRef.SizeBytes
		}
	}
	return size
}

func newDataRef(chunkRef *chunk.Ref, offset, size int64) *chunk.DataRef {
	return &chunk.DataRef{
		Ref:         chunkRef,
		OffsetBytes: offset,
		SizeBytes:   size,
	}
}
