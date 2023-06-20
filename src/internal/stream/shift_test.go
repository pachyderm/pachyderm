package stream

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestBlueshift(t *testing.T) {
	ctx := pctx.TestContext(t)
	it1 := NewSlice([][]int{
		{0, 1, 2, 3},
		{4, 5, 6},
		{7, 8},
		{9, 10},
	})
	it2 := Blueshift[int](it1, DefaultCopy[int])

	xs, err := Collect(ctx, it2, 100)
	require.NoError(t, err)
	for i, x := range xs {
		require.Equal(t, i, x)
	}
}
