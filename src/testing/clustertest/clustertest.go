package clustertest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// TestAPI is a test suite which checks that objects correctly implement the pachyderm APIs.
func TestAPI(t *testing.T, newClient func(t testing.TB) *client.APIClient) {
	tests := []struct {
		Name string
		F    func(t *testing.T, c *client.APIClient)
	}{
		{"SimplePipeline", testSimplePipeline},
		// Add more tests here
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			cluster := newClient(t)
			test.F(t, cluster)
		})
	}
}

func testSimplePipeline(t *testing.T, c *client.APIClient) {
	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(dataRepo, commit1.Branch.Name, commit1.ID))

	pipeline := tu.UniqueString("TestSimplePipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Branch.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(dataRepo))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Branch.Repo, client.NewRepo(pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}
