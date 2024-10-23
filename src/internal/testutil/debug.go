package testutil

import (
	"path"
	"testing"

	globlib "github.com/pachyderm/ohmyglob"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

func DebugFiles(t testing.TB, projectName, repoName string) (map[string]*globlib.Glob, []string) {
	expectedFiles := make(map[string]*globlib.Glob)
	// Record glob patterns for expected pachd files.
	for _, file := range []string{"version.txt", "logs.txt", "logs-previous**", "logs-loki.txt", "goroutine", "heap"} {
		pattern := path.Join("pachd", "**", "pachd", file)
		g, err := globlib.Compile(pattern, '/')
		require.NoError(t, err)
		expectedFiles[pattern] = g
	}
	pattern := path.Join("pachd", "**", "describe.txt")
	g, err := globlib.Compile(pattern, '/')
	require.NoError(t, err)
	expectedFiles[pattern] = g
	// Record glob patterns for expected source repo files.
	for _, file := range []string{"commits.json", "commits-chart**"} {
		pattern := path.Join("source-repos", projectName, repoName, file)
		g, err := globlib.Compile(pattern, '/')
		require.NoError(t, err)
		expectedFiles[pattern] = g
	}
	var pipelines []string
	for i := 0; i < 3; i++ {
		pipeline := uuid.UniqueString("TestDebug")
		pipelines = append(pipelines, pipeline)
		// Record glob patterns for expected pipeline files.
		pattern := path.Join("pipelines", projectName, pipeline, "pods", "*", "describe.txt")
		g, err := globlib.Compile(pattern, '/')
		require.NoError(t, err)
		expectedFiles[pattern] = g
		for _, container := range []string{"user", "storage"} {
			for _, file := range []string{"logs.txt", "logs-previous**", "logs-loki.txt", "goroutine", "heap"} {
				pattern := path.Join("pipelines", projectName, pipeline, "pods", "*", container, file)
				g, err := globlib.Compile(pattern, '/')
				require.NoError(t, err)
				expectedFiles[pattern] = g
			}
		}
		for _, file := range []string{"spec.json", "commits.json", "jobs.json", "commits-chart**", "jobs-chart**"} {
			pattern := path.Join("pipelines", projectName, pipeline, file)
			g, err := globlib.Compile(pattern, '/')
			require.NoError(t, err)
			expectedFiles[pattern] = g
		}
	}
	for _, app := range []string{"etcd", "pg-bouncer"} {
		for _, file := range []string{"describe.txt", "logs.txt", "logs-previous**", "logs-loki.txt"} {
			pattern := path.Join(app, "**", file)
			g, err := globlib.Compile(pattern, '/')
			require.NoError(t, err)
			expectedFiles[pattern] = g
		}
	}
	return expectedFiles, pipelines
}
