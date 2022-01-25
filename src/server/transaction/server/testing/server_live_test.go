//go:build livek8s
// +build livek8s

package testing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestCreatePipelineTransaction(t *testing.T) {
	c := testutil.GetPachClient(t)
	require.NoError(t, c.DeleteAll())
	repo := testutil.UniqueString("in")
	pipeline := testutil.UniqueString("pipeline")
	_, err := c.ExecuteInTransaction(func(txnClient *client.APIClient) error {
		require.NoError(t, txnClient.CreateRepo(repo))
		require.NoError(t, txnClient.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out", repo)},
			&pps.ParallelismSpec{Constant: 1},
			client.NewPFSInput(repo, "/"),
			"master",
			false,
		))
		return nil
	})
	require.NoError(t, err)

	commit := client.NewCommit(repo, "master", "")
	c.PutFile(commit, "foo", strings.NewReader("bar"))

	commitInfo, err := c.WaitCommit(pipeline, "master", "")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfo.Commit, "foo", &buf))
	require.Equal(t, "bar", buf.String())
}
