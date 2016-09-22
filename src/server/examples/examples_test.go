package examples

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}

func TestExampleTensorFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	exampleDir := filepath.Join(cwd, "../../../examples/tensor_flow")
	cmd := exec.Command("make", "test")
	cmd.Dir = exampleDir
	raw, err := cmd.CombinedOutput()
	fmt.Printf("make all output: %v\n", string(raw))
	require.NoError(t, err)

	commitInfos, err := c.ListCommit([]string{"GoT_scripts"}, nil, client.CommitTypeRead, false, client.CommitStatusAll, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	inputCommitID := commitInfos[0].Commit.ID

	// Wait until the GoT_generate job has finished
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit("GoT_scripts", inputCommitID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	repos := []interface{}{"GoT_train", "GoT_generate"}
	var generateCommitID string
	for _, commitInfo := range commitInfos {
		require.EqualOneOf(t, repos, commitInfo.Commit.Repo.Name)
		if commitInfo.Commit.Repo.Name == "GoT_generate" {
			generateCommitID = commitInfo.Commit.ID
		}
	}

	// Make sure the final output is non zero
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile("GoT_generate", generateCommitID, "new_script.txt", 0, 0, "", false, nil, &buffer))
	if buffer.Len() < 100 {
		t.Fatalf("Output GoT script is too small (has len=%v)", buffer.Len())
	}
	require.NoError(t, c.DeleteRepo("GoT_generate", false))
	require.NoError(t, c.DeleteRepo("GoT_train", false))
	require.NoError(t, c.DeleteRepo("GoT_scripts", false))
}
