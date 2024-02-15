//go:build k8s

package datum

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func BenchmarkPipeline(b *testing.B) {
	c, _ := minikubetestenv.AcquireCluster(b)

	b.ResetTimer()
	require.NoError(b, c.CreateRepo(pfs.DefaultProjectName, "input"))
	commit, err := c.StartCommit(pfs.DefaultProjectName, "input", "master")
	require.NoError(b, err, "should be able to create commit")
	for i := 0; i < b.N; i++ {
		for j := 0; j < 125; j++ {
			require.NoError(b, c.PutFile(commit, fmt.Sprintf("/file%d-%d", i, j), strings.NewReader("test content here")))
		}
	}
	require.NoError(b, c.CreatePipeline(pfs.DefaultProjectName, "first", "", []string{"/bin/bash"}, []string{"cp /pfs/input/* /pfs/out"}, &pps.ParallelismSpec{Constant: 8}, &pps.Input{Pfs: &pps.PFSInput{Glob: "/*", Repo: "input"}}, "master", false))
	require.NoError(b, c.FinishCommit(pfs.DefaultProjectName, "input", "master", commit.Id))
}

func BenchmarkIncrementalPipeline(b *testing.B) {
	c, _ := minikubetestenv.AcquireCluster(b)
	require.NoError(b, c.CreateRepo(pfs.DefaultProjectName, "input"))
	commit, err := c.StartCommit(pfs.DefaultProjectName, "input", "master")
	require.NoError(b, err, "should be able to create commit")
	txn, err := c.StartTransaction()
	require.NoError(b, err, "should be able to create transaction")
	for i := 0; i < b.N; i++ {
		for j := 0; j < 125; j++ {
			require.NoError(b, c.PutFile(commit, fmt.Sprintf("/file%d-%d", i, j), strings.NewReader("new test content here")))
		}
	}
	require.NoError(b, c.CreatePipeline(pfs.DefaultProjectName, "first", "", []string{"/bin/bash"}, []string{"cp /pfs/input/* /pfs/out"}, &pps.ParallelismSpec{Constant: 8}, &pps.Input{Pfs: &pps.PFSInput{Glob: "/*", Repo: "input"}}, "master", false))
	require.NoError(b, c.FinishCommit(pfs.DefaultProjectName, "input", "master", commit.Id))
	_, err = c.FinishTransaction(txn)
	require.NoError(b, err, "should finish transaction")

	b.ResetTimer()

	commit, err = c.StartCommit(pfs.DefaultProjectName, "input", "master")
	require.NoError(b, err, "should be able to create commit")
	txn, err = c.StartTransaction()
	require.NoError(b, err, "should be able to create transaction")
	for i := 0; i < b.N; i++ {
		require.NoError(b, c.PutFile(commit, fmt.Sprintf("/file%d-%d", i, i), strings.NewReader("incremental test content here")))
	}
	require.NoError(b, c.FinishCommit(pfs.DefaultProjectName, "input", "master", commit.Id))
	_, err = c.FinishTransaction(txn)
	require.NoError(b, err, "should finish transaction")
}
