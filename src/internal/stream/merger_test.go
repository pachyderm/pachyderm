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
	expected := []string{"a", "b", "c", "d"}

	m := NewMerger(its, func(a, b string) bool {
		return a < b
	})
	actual, err := Collect[string](pctx.TestContext(t), m, 100)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
