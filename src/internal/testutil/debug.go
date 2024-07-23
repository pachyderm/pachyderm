package testutil

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	globlib "github.com/pachyderm/ohmyglob"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
		pipeline := UniqueString("TestDebug")
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

const (
	CommitsTable  = "pfs/commits"
	BranchesTable = "pfs/branches"
	ReposTable    = "pfs/repos"
	ProjectsTable = "pfs/projects"
)

func Dump(t testing.TB, c *client.APIClient) map[string]string {
	filter := &debug.Filter{
		Filter: &debug.Filter_Database{Database: true},
	}
	buf := &bytes.Buffer{}
	time.Sleep(5 * time.Second) // give some time for the stats collector to run.
	require.NoError(t, c.Dump(filter, 100, buf), "dumping database files should succeed")
	gr, err := gzip.NewReader(buf)
	require.NoError(t, err)

	db := make(map[string]string)

	require.NoError(t, tarutil.Iterate(gr, func(f tarutil.File) error {
		fileContents := &bytes.Buffer{}
		if err := f.Content(fileContents); err != nil {
			return errors.EnsureStack(err)
		}
		hdr, err := f.Header()
		require.NoError(t, err, "getting database tar file header should succeed")
		if strings.HasPrefix(hdr.Name, "database/tables/") {
			db[strings.TrimSuffix(strings.TrimPrefix(hdr.Name, "database/tables/"), ".json")] = fileContents.String()
		}
		return nil
	}))
	return db
}

func DumpDatabaseCommitsOverTime(t testing.TB, c *client.APIClient, repo *pfs.Repo, branchName string, from string, state pfs.CommitState) {
	go func() {
		var lastDump []map[string]interface{}
		require.NoError(t, c.SubscribeCommit(repo, branchName, from, state, func(ci *pfs.CommitInfo) error {
			dump := (Dump(t, c))[CommitsTable]
			var currDump []map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(dump), &currDump))
			sort.Slice(currDump, func(i, j int) bool {
				return currDump[i]["int_id"].(float64) < currDump[j]["int_id"].(float64)
			})
			diff := cmp.Diff(lastDump, currDump)
			if diff != "" {
				t.Log(diff)
			}
			lastDump = currDump
			return nil
		}))
	}()

}
