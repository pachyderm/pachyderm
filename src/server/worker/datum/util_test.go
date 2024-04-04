package datum

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCartesianProduct(t *testing.T) {
	inputsShards := [][]string{
		{"a", "b", "c"},
		{"1", "2"},
		{"x", "y"},
	}

	permutations := cartesianProduct(inputsShards, 0)
	require.Equal(t, 4, len(permutations))
	expected := [][]string{
		{"c", "1", "x"},
		{"c", "1", "y"},
		{"c", "2", "x"},
		{"c", "2", "y"},
	}
	for _, e := range expected {
		require.True(t, sliceExistsInSlices(e, permutations))
	}

	permutations = cartesianProduct(inputsShards, 1)
	require.Equal(t, 6, len(permutations))
	expected = [][]string{
		{"a", "2", "x"},
		{"a", "2", "y"},
		{"b", "2", "x"},
		{"b", "2", "y"},
		{"c", "2", "x"},
		{"c", "2", "y"},
	}
	for _, e := range expected {
		require.True(t, sliceExistsInSlices(e, permutations))
	}
}

func slicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[string]int)
	for _, v := range a {
		count[v]++
	}
	for _, v := range b {
		count[v]--
		if count[v] < 0 {
			return false
		}
	}
	return true
}

func sliceExistsInSlices(s []string, slices [][]string) bool {
	for _, slice := range slices {
		if slicesEqualUnordered(s, slice) {
			return true
		}
	}
	return false
}
