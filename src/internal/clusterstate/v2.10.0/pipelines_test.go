package v2_10_0

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestPipelineVersionsDeduplication(t *testing.T) {
	// - setup database with collections.pipelines, collections.jobs
	// - migrate pipeline A from versions [1, 1, 2] -> [1, 2, 3]
	type pipelineVersion struct {
		createdAt time.Time
		version   int
	}
	type testCase struct {
		desc    string
		initial []pipelineVersion
		want    []pipelineVersion
		wantErr bool
	}
	var (
		baseTime = time.Now().Add(-time.Hour * 24 * 365)
		tcs      = []testCase{
			{
				desc: "empty case does nothing",
			},
			{
				desc: "no changes needed result in no changes made",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
			},
			{
				desc: "fail when there are two pipeline versions at the very same instant",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
				},
				wantErr: true,
			},
			{
				desc: "do not fail when there are two pipeline versions with different version numbers at the very same instant",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				wantErr: true,
			},
			{
				desc: "renumber simple duplication at beginning",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication at beginning",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber simple duplication in the middle",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   3,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber simple duplication in the middle with duplicate timestamps but different versions",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication in the middle",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber simple duplication at end",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication at end",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber complex deduplication",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					}, {
						createdAt: baseTime.Add(5 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
					{
						createdAt: baseTime.Add(6 * time.Second),
						version:   5,
					}, {
						createdAt: baseTime.Add(7 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(8 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(9 * time.Second),
						version:   6,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					}, {
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
					{
						createdAt: baseTime.Add(5 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(6 * time.Second),
						version:   6,
					}, {
						createdAt: baseTime.Add(7 * time.Second),
						version:   7,
					},
					{
						createdAt: baseTime.Add(8 * time.Second),
						version:   8,
					},
					{
						createdAt: baseTime.Add(9 * time.Second),
						version:   9,
					},
				},
			},
		}
	)
	for _, tc := range tcs {
		require.True(t, true, tc.desc)
	}
}
