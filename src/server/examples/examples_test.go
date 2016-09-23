package examples

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	exampleDir := filepath.Join(cwd, "../../../doc/examples/tensor_flow")
	cmd := exec.Command("make", "test")
	cmd.Dir = exampleDir
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	commitInfos, err := c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"GoT_scripts"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusAll,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	inputCommitID := commitInfos[0].Commit.ID

	// Wait until the GoT_generate job has finished
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit("GoT_scripts", inputCommitID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	repos := []interface{}{"GoT_train", "GoT_generate", "GoT_scripts"}
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

func TestWordCount(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)

	// Should stay in sync with doc/examples/word_count/README.md
	inputPipelineManifest := `
{
  "pipeline": {
    "name": "wordcount_input"
  },
  "transform": {
    "image": "pachyderm/job-shim:latest",
    "cmd": [ "wget",
        "-e", "robots=off",
        "--recursive",
        "--level", "1",
        "--adjust-extension",
        "--no-check-certificate",
        "--no-directories",
        "--directory-prefix",
        "/pfs/out",
        "https://news.ycombinator.com/newsfaq.html"
    ],
    "acceptReturnCode": [4,5,6,7,8]
  },
  "parallelism_spec": {
       "strategy" : "CONSTANT",
       "constant" : 1
  }
}
`
	exampleDir := "../../../doc/examples/word_count"
	cmd := exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(inputPipelineManifest)
	cmd.Dir = exampleDir
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "run-pipeline", "wordcount_input")
	cmd.Dir = exampleDir
	_, err = cmd.Output()
	require.NoError(t, err)

	cmd = exec.Command("docker", "build", "-t", "wordcount-map", ".")
	cmd.Dir = exampleDir
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	// Should stay in sync with doc/examples/word_count/README.md
	wordcountMapPipelineManifest := `
{
  "pipeline": {
    "name": "wordcount_map"
  },
  "transform": {
    "image": "wordcount-map:latest",
    "cmd": ["/map", "/pfs/wordcount_input", "/pfs/out"]
  },
  "inputs": [
    {
      "repo": {
        "name": "wordcount_input"
      }
    }
  ]
}
`
	cmd = exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(wordcountMapPipelineManifest)
	cmd.Dir = exampleDir
	raw, err = cmd.Output()
	require.NoError(t, err)

	// Flush Commit can't help us here since there are no inputs
	// So we poll wordcount_input until it has a commit
	tries := 10
	var commitInfos []*pfsclient.CommitInfo
	for tries != 0 {
		commitInfos, err = c.ListCommit(
			[]*pfsclient.Commit{{
				Repo: &pfsclient.Repo{"wordcount_input"},
			}},
			nil,
			client.CommitTypeRead,
			client.CommitStatusNormal,
			false,
		)
		require.NoError(t, err)
		if len(commitInfos) == 1 {
			break
		}
		time.Sleep(10 * time.Second)
		tries--
	}
	require.Equal(t, 1, len(commitInfos))
	inputCommit := commitInfos[0].Commit
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{inputCommit}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"wordcount_map"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusNormal,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "are", 0, 0, "", false, nil, &buffer))
	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	// Should see # lines output == # pods running job
	// This should be just one with default deployment
	require.Equal(t, 1, len(lines))

	// Should stay in sync with doc/examples/word_count/README.md
	wordcountReducePipelineManifest := `
{
  "pipeline": {
    "name": "wordcount_reduce"
  },
  "transform": {
    "image": "pachyderm/job-shim:latest",
    "cmd": ["sh"],
    "stdin": [
        "find /pfs/wordcount_map -name '*' | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count /pfs/out/` + "`basename $count`" + `; done"
    ]
  },
  "inputs": [
    {
      "repo": {
        "name": "wordcount_map"
      },
	  "method": "reduce"
    }
  ]
}
`

	cmd = exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(wordcountReducePipelineManifest)
	cmd.Dir = exampleDir
	_, err = cmd.Output()
	require.NoError(t, err)

	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{inputCommit}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"wordcount_reduce"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusNormal,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	buffer.Reset()
	require.NoError(t, c.GetFile("wordcount_reduce", commitInfos[0].Commit.ID, "morning", 0, 0, "", false, nil, &buffer))
	lines = strings.Split(buffer.String(), "\n")
	require.Equal(t, 1, len(lines))

	fileInfos, err := c.ListFile("wordcount_reduce", commitInfos[0].Commit.ID, "", "", false, nil, false)
	require.NoError(t, err)

	if len(fileInfos) < 100 {
		t.Fatalf("Word count result is too small. Should have counted a bunch of words. Only counted %v:\n%v\n", len(fileInfos), fileInfos)
	}
}
