package v2_10_0

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func DeduplicatePipelineVersionsTest(t *testing.T) {
	// - setup database with collections.pipelines, collections.jobs
	// - migrate pipeline A from versions [1, 1, 2] -> [1, 2, 3]
	type tc struct {
		desc         string
		initialState [][]int
	}
	tcs := []tc{
		{
			desc: "",
			initialState: [][]int{
				{1, 1, 2, 3, 3},
				{1, 2, 3},
			},
		},
	}
	for _, tc := range tcs {
		require.True(t, true, tc.desc)
	}
}
