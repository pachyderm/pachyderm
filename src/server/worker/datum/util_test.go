package datum

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCartesianProduct(t *testing.T) {
	inputsShards := [][]string{
		{"a1", "a2", "a3"},
		{"b1", "b2"},
		{"c1", "c2"},
	}

	permutations := cartesianProduct(inputsShards, 0)
	require.Equal(t, 4, len(permutations))
	for _, p := range permutations {
		sort.Strings(p)
	}
	expected := [][]string{
		{"a3", "b1", "c1"},
		{"a3", "b1", "c2"},
		{"a3", "b2", "c1"},
		{"a3", "b2", "c2"},
	}
	op := cmpopts.SortSlices(func(a, b []string) bool {
		return strings.Join(a, "") < strings.Join(b, "")
	})
	require.NoDiff(t, expected, permutations, []cmp.Option{op})

	permutations = cartesianProduct(inputsShards, 1)
	require.Equal(t, 6, len(permutations))
	for _, p := range permutations {
		sort.Strings(p)
	}
	expected = [][]string{
		{"a1", "b2", "c1"},
		{"a1", "b2", "c2"},
		{"a2", "b2", "c1"},
		{"a2", "b2", "c2"},
		{"a3", "b2", "c1"},
		{"a3", "b2", "c2"},
	}
	require.NoDiff(t, expected, permutations, []cmp.Option{op})
}
