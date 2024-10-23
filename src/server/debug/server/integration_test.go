package server_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func createTestCommits(t *testing.T, repoName, branchName string, numCommits int, c *client.APIClient) {
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repoName), "pods should be available to create a repo against")

	for i := 0; i < numCommits; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, repoName, branchName)
		require.NoError(t, err, "creating a test commit should succeed")

		err = c.PutFile(commit, "/file", bytes.NewBufferString("file contents"))
		require.NoError(t, err, "should be able to add a file to a commit")

		err = c.FinishCommit(pfs.DefaultProjectName, repoName, branchName, commit.Id)
		require.NoError(t, err, "finishing a commit should succeed")
	}
}

func TestDatabaseStats(t *testing.T) {
	type rowCountResults struct {
		NLiveTup   int    `json:"n_live_tup"`
		RelName    string `json:"relname"`
		SchemaName string `json:"schemaname"`
	}

	type commitResults struct {
		Key string `json:"key"`
	}

	repoName := uuid.UniqueString("TestDatabaseStats-repo")
	branchName := "master"
	numCommits := 100
	buf := &bytes.Buffer{}
	filter := &debug.Filter{
		Filter: &debug.Filter_Database{Database: true},
	}

	c := pachd.NewTestPachd(t)

	createTestCommits(t, repoName, branchName, numCommits, c)
	time.Sleep(5 * time.Second) // give some time for the stats collector to run.
	require.NoError(t, c.Dump(filter, 100, buf), "dumping database files should succeed")
	gr, err := gzip.NewReader(buf)
	require.NoError(t, err)

	var rows []commitResults
	foundCommitsJSON, foundRowCounts := false, false
	require.NoError(t, tarutil.Iterate(gr, func(f tarutil.File) error {
		fileContents := &bytes.Buffer{}
		if err := f.Content(fileContents); err != nil {
			return errors.EnsureStack(err)
		}

		hdr, err := f.Header()
		require.NoError(t, err, "getting database tar file header should succeed")
		require.NotMatch(t, "^[a-zA-Z0-9_\\-\\/]+\\.error$", hdr.Name)
		switch hdr.Name {
		case "database/row-counts.json":
			var rows []rowCountResults
			require.NoError(t, json.Unmarshal(fileContents.Bytes(), &rows),
				"unmarshalling row-counts.json should succeed")

			for _, row := range rows {
				if row.RelName == "commits" && row.SchemaName == "pfs" {
					require.NotEqual(t, 0, row.NLiveTup,
						"some commits from createTestCommits should be accounted for")
					foundRowCounts = true
				}
			}
		case "database/tables/pfs/commits.json":
			require.NoError(t, json.Unmarshal(fileContents.Bytes(), &rows),
				"unmarshalling commits.json should succeed")
			require.Equal(t, numCommits, len(rows), "number of commits should match number of rows")
			foundCommitsJSON = true
		}

		return nil
	}))

	require.Equal(t, true, foundRowCounts,
		"we should have an entry in row-counts.json for commits")
	require.Equal(t, true, foundCommitsJSON,
		"checks for commits.json should succeed")
}
