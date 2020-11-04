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

// IndexPointsTo returns a list of all the chunks this index references
func PointsTo(idx *Index) (ids []chunk.ChunkID) {
	if idx == nil || idx.DataOp == nil {
		return nil
	}
	for _, dr := range idx.DataOp.DataRefs {
		ids = append(ids, chunk.ChunkID(dr.ChunkRef.Id))
	}
	return ids
}
