package examples

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
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

func toRepoNames(pfsRepos []*pfsclient.RepoInfo) []interface{} {
	repoNames := make([]interface{}, len(pfsRepos))
	for _, repoInfo := range pfsRepos {
		repoNames = append(repoNames, repoInfo.Repo.Name)
	}
	return repoNames
}

func TestWordCount(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	exampleDir := "../../../examples/word_count"
	newURL := "https://news.ycombinator.com/newsfaq.html"
	oldURL := "https://en.wikipedia.org/wiki/Main_Page"
	rawInputPipelineManifest, err := ioutil.ReadFile(filepath.Join(exampleDir, "inputPipeline.json"))
	require.NoError(t, err)
	// Need to choose a page w much fewer links to make this pass on CI
	inputPipelineManifest := strings.Replace(string(rawInputPipelineManifest), oldURL, newURL, 1)

	cmd := exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(inputPipelineManifest)
	cmd.Dir = exampleDir
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "run-pipeline", "input")
	cmd.Dir = exampleDir
	_, err = cmd.Output()
	require.NoError(t, err)

	cmd = exec.Command("docker", "build", "-t", "wordcount-map", ".")
	cmd.Dir = exampleDir
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	wordcountMapPipelineManifest, err := ioutil.ReadFile(filepath.Join(exampleDir, "mapPipeline.json"))
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(string(wordcountMapPipelineManifest))
	cmd.Dir = exampleDir
	_, err = cmd.Output()
	require.NoError(t, err)

	// Flush Commit can't help us here since there are no inputs
	// So we block on ListCommit
	commitInfos, err := c.ListCommitByRepo("input")
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	inputCommit := commitInfos[0].Commit
	iter, err := c.FlushCommit([]*pfsclient.Commit{inputCommit}, nil)
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		info, err := iter.Next()
		require.NotNil(t, info)
		require.NoError(t, err)
	}

	commitInfos, err = c.ListCommitByRepo("map")
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	// Pick a random word and check its content
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "/are", 0, 0, &buffer))
	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	// Should see # lines output == parallelism (each worker outputs one line, and
	// all lines get appended). This should be just one with default deployment
	require.Equal(t, 1, len(lines))

	wordcountReducePipelineManifest, err := ioutil.ReadFile(filepath.Join(exampleDir, "reducePipeline.json"))
	require.NoError(t, err)

	cmd = exec.Command("pachctl", "create-pipeline")
	cmd.Stdin = strings.NewReader(string(wordcountReducePipelineManifest))
	cmd.Dir = exampleDir
	_, err = cmd.Output()
	require.NoError(t, err)

	iter, err = c.FlushCommit([]*pfsclient.Commit{inputCommit}, nil)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		info, err := iter.Next()
		require.NotNil(t, info)
		require.NoError(t, err)
	}

	commitInfos, err = c.ListCommitByRepo("reduce")
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	buffer.Reset()
	require.NoError(t, c.GetFile("reduce", commitInfos[0].Commit.ID, "/morning", 0, 0, &buffer))
	lines = strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	require.Equal(t, 1, len(lines))

	fileInfos, err := c.ListFile("reduce", commitInfos[0].Commit.ID, "/")
	require.NoError(t, err)
	if len(fileInfos) < 100 {
		t.Fatalf("Word count result is too small. Should have counted a bunch of words. Only counted %v:\n%v\n", len(fileInfos), fileInfos)
	}
}
