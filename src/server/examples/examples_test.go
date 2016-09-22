package examples

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func TestFruitStand(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	t.Parallel()

	require.NoError(t, c.CreateRepo("data"))
	repoInfos, err := c.ListRepo(nil)
	require.NoError(t, err)
	var repoNames []interface{}
	for _, repoInfo := range repoInfos {
		repoNames = append(repoNames, repoInfo.Repo.Name)
	}
	require.OneOfEquals(t, "data", repoNames)

	cmd := exec.Command(
		"pachctl",
		"put-file",
		"data",
		"master",
		"sales",
		"-c",
		"-f",
		"https://raw.githubusercontent.com/pachyderm/pachyderm/master/doc/examples/fruit_stand/set1.txt",
	)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	commitInfos, err := c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"data"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusNormal,
		false,
	)
	require.Equal(t, 1, len(commitInfos))
	commit := commitInfos[0].Commit

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile("data", commit.ID, "sales", 0, 0, "", false, nil, &buffer))
	lines := strings.Split(buffer.String(), "\n")
	if len(lines) < 100 {
		t.Fatalf("Sales file has too few lines (%v)\n", len(lines))
	}

	cmd = exec.Command(
		"pachctl",
		"create-pipeline",
		"-f",
		"https://raw.githubusercontent.com/pachyderm/pachyderm/master/doc/examples/fruit_stand/pipeline.json",
	)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	time.Sleep(5)
	repoInfos, err = c.ListRepo(nil)
	require.NoError(t, err)
	repoNames = []interface{}{}
	for _, repoInfo := range repoInfos {
		repoNames = append(repoNames, repoInfo.Repo.Name)
	}
	require.OneOfEquals(t, "sum", repoNames)
	require.OneOfEquals(t, "filter", repoNames)
	require.OneOfEquals(t, "data", repoNames)

	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit("data", commit.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"sum"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusNormal,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	buffer.Reset()
	require.NoError(t, c.GetFile("sum", commitInfos[0].Commit.ID, "apple", 0, 0, "", false, nil, &buffer))
	require.NotNil(t, buffer)
	firstCount, err := strconv.ParseInt(strings.TrimRight(buffer.String(), "\n"), 10, 0)
	require.NoError(t, err)
	if firstCount < 100 {
		t.Fatalf("Wrong sum for apple (%i) ... too low\n", firstCount)
	}

	// Add more data to input

	cmd = exec.Command(
		"pachctl",
		"put-file",
		"data",
		"master",
		"sales",
		"-c",
		"-f",
		"https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/fruit_stand/set2.txt",
	)
	_, err = cmd.Output()
	require.NoError(t, err)

	// Flush commit!
	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"data"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusAll,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit("data", commitInfos[1].Commit.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{"sum"},
		}},
		nil,
		client.CommitTypeRead,
		client.CommitStatusNormal,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	buffer.Reset()
	require.NoError(t, c.GetFile("sum", commitInfos[1].Commit.ID, "apple", 0, 0, "", false, nil, &buffer))
	require.NotNil(t, buffer)
	secondCount, err := strconv.ParseInt(strings.TrimRight(buffer.String(), "\n"), 10, 0)
	require.NoError(t, err)
	if firstCount > secondCount {
		t.Fatalf("Second sum (%v) is smaller than first (%v)\n", secondCount, firstCount)
	}

	require.NoError(t, c.DeleteRepo("sum", false))
	require.NoError(t, c.DeleteRepo("filter", false))
	require.NoError(t, c.DeleteRepo("data", false))
}
