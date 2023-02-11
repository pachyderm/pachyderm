package stream

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMerger(t *testing.T) {
	its := []Peekable[string]{
		NewSlice([]string{"c", "d"}),
		NewSlice([]string{"a"}),
		NewSlice([]string{"b", "c"}),
		NewSlice([]string{}),
	}

	expected := []Merged[string]{
		{Indexes: []int{1}, Values: []string{"a"}},
		{Indexes: []int{2}, Values: []string{"b"}},
		{Indexes: []int{0, 2}, Values: []string{"c", "c"}},
		{Indexes: []int{0}, Values: []string{"d"}},
	}
	m := NewMerger(its, func(a, b string) bool {
		return a < b
	})

	actual, err := Collect[Merged[string]](pctx.TestContext(t), m, 100)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestReducer(t *testing.T) {
	its := []Peekable[string]{
		NewSlice([]string{"c", "d"}),
		NewSlice([]string{"a"}),
		NewSlice([]string{"b", "c"}),
		NewSlice([]string{}),
	}
	expected := []string{"a", "b", "c", "d"}

	m := NewReducer(its, func(a, b string) bool {
		return a < b
	}, func(dst *string, m Merged[string]) {
		*dst = m.Values[len(m.Values)-1]
	})
	actual, err := Collect[string](pctx.TestContext(t), m, 100)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
