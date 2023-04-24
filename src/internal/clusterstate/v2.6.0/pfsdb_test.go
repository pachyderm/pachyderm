//go:build unit_test

package v2_6_0

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	v2_5_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.5.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestValidateOldDAGs(t *testing.T) {
	type testCase struct {
		cis       []*v2_5_0.CommitInfo
		expectErr bool
	}
	makeCommit := func(c *pfs.Commit, parent *pfs.Commit) *v2_5_0.CommitInfo {
		return &v2_5_0.CommitInfo{
			Commit:       c,
			ParentCommit: parent,
		}
	}
	p := pfs.DefaultProjectName
	r1 := "repo1"
	r2 := "repo2"
	b1 := "branch1"
	b2 := "branch2"
	id1 := "abc"
	cases := []testCase{
		{
			cis: []*v2_5_0.CommitInfo{
				makeCommit(client.NewProjectCommit(p, r1, b1, id1), nil),
				makeCommit(client.NewProjectCommit(p, r2, b1, id1), nil),
			},
			expectErr: false,
		},
		{
			cis: []*v2_5_0.CommitInfo{
				makeCommit(client.NewProjectCommit(p, r1, b1, id1), nil),
				makeCommit(client.NewProjectCommit(p, r1, b2, id1), client.NewProjectCommit(p, r1, b1, id1)),
			},
			expectErr: false,
		},
		{
			cis: []*v2_5_0.CommitInfo{
				makeCommit(client.NewProjectCommit(p, r1, b1, id1), nil),
				makeCommit(client.NewProjectCommit(p, r1, b2, id1), nil),
			},
			expectErr: true,
		},
	}
	for _, c := range cases {
		err := validateExistingDAGsHelper(c.cis)
		if c.expectErr {
			require.YesError(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
