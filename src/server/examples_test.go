package server

import (
	"fmt"
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
	fmt.Printf("make output: %v\n", string(raw))
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "list-commit", "GoT_scripts")
	cmd.Dir = exampleDir
	raw, err = cmd.CombinedOutput()
	fmt.Printf("Raw output: %v\n", string(raw))
	lines := strings.Split(string(raw), "\n")
	fmt.Printf("Lines: %v\n", lines)
	require.Equal(t, 2, len(lines))

	getSecondField := func(line string) string {
		tokens := strings.Split(line, " ")
		seenField := 0
		for _, token := range tokens {
			if seenField != 0 && token != "" {
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

	// Wait until the GoT/generate job has finished
	c := getPachClient(t)
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit("GoT_scripts", inputCommitID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	fmt.Printf("Commits flushed: %v\n", commitInfos)

	// Make sure the final output is non zero

}
