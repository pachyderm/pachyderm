package server

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestTensorFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)
	exampleDir := filepath.Join(cwd, "../../examples/tensor_flow")
	cmd := exec.Command("make", "all")
	cmd.Dir = exampleDir
	raw, err := cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "list-commit", "GoT_scripts")
	cmd.Dir = exampleDir
	raw, err = cmd.CombinedOutput()
	lines := strings.Split(string(raw), "\n")
	require.Equal(t, 3, len(lines))

	getSecondField := func(line string) string {
		tokens := strings.Split(line, " ")
		seenField := 0
		for _, token := range tokens {
			if token != "" {
				seenField += 1
				if seenField == 2 {
					return token
				}
			}
		}
		return ""
	}

	// Make sure the second field is `ID`
	// Example stdout we're parsing:
	//
	// BRANCH                             ID                                 PARENT              STARTED             FINISHED            SIZE
	// c1001a97c0cc4bea825ee50cc613e039   5fc5a07edd094432acf474662ad02854   <none>              About an hour ago   About an hour ago   2.625 MiB

	require.Equal(t, "ID", getSecondField(lines[0]))
	inputCommitID := getSecondField(lines[1])
	require.NotEqual(t, "", inputCommitID)

	// Wait until the GoT_generate job has finished
	c := getPachClient(t)
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit("GoT_scripts", inputCommitID)}, nil)
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
	require.NoError(t, c.DeleteAll())
}
