package index

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
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

func resolveDataOps(idx *Index) {
	if idx.FileOp.DataRefs == nil {
		return
	}
	dataRefs := idx.FileOp.DataRefs
	offset := dataRefs[0].OffsetBytes
	size := dataRefs[0].SizeBytes
	for _, dataOp := range idx.FileOp.DataOps {
		bytesLeft := dataOp.SizeBytes
		for size <= bytesLeft {
			dataOp.DataRefs = append(dataOp.DataRefs, newDataRef(dataRefs[0].ChunkInfo, offset, size))
			bytesLeft -= size
			dataRefs = dataRefs[1:]
			if len(dataRefs) == 0 {
				return
			}
			offset = dataRefs[0].OffsetBytes
			size = dataRefs[0].SizeBytes
		}
		dataOp.DataRefs = append(dataOp.DataRefs, newDataRef(dataRefs[0].ChunkInfo, offset, bytesLeft))
		offset += bytesLeft
		size -= bytesLeft
	}
}

func newDataRef(chunkInfo *chunk.ChunkInfo, offset, size int64) *chunk.DataRef {
	return &chunk.DataRef{
		ChunkInfo:   chunkInfo,
		OffsetBytes: offset,
		SizeBytes:   size,
	}
}

func unresolveDataOps(idx *Index) {
	for _, dataOp := range idx.FileOp.DataOps {
		dataOp.DataRefs = nil
	}
}
