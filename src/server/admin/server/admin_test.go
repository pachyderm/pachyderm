package server

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	versionlib "github.com/pachyderm/pachyderm/src/client/version"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

const (
	KB = 1024
	MB = 1024 * KB
)

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		require.NoError(t, err)
	})
	return pachClient
}

func collectCommitInfos(t testing.TB, commitInfoIter client.CommitInfoIterator) []*pfs.CommitInfo {
	var commitInfos []*pfs.CommitInfo
	for {
		commitInfo, err := commitInfoIter.Next()
		if err == io.EOF {
			return commitInfos
		}
		require.NoError(t, err)
		commitInfos = append(commitInfos, commitInfo)
	}
}

func TestExtractRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	nCommits := 2
	r := rand.New(rand.NewSource(45))
	fileContent := workload.RandString(r, 40*MB)
	for i := 0; i < nCommits; i++ {
		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(fileContent))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, "master"))
	}

	numPipelines := 3
	input := dataRepo
	for i := 0; i < numPipelines; i++ {
		pipeline := tu.UniqueString(fmt.Sprintf("TestExtractRestore%d", i))
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(input, "/*"),
			"",
			false,
		))
		input = pipeline
	}

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))

	commitIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))
}

func TestExtractRestoreHeadlessBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create a headless branch
	require.NoError(t, c.CreateBranch(dataRepo, "headless", "", nil))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))

	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))
	require.Equal(t, "headless", bis[0].Branch.Name)
}

func TestExtractVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	r := rand.New(rand.NewSource(45))
	_, err := c.PutFile(dataRepo, "master", "file", strings.NewReader(workload.RandString(r, 40*MB)))
	require.NoError(t, err)

	pipeline := tu.UniqueString("TestExtractRestore")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.True(t, len(ops) > 0)

	// Check that every Op looks right; the version set matches pachd's version
	for _, op := range ops {
		opV := reflect.ValueOf(op).Elem()
		var versions, nonemptyVersions int
		for i := 0; i < opV.NumField(); i++ {
			fDesc := opV.Type().Field(i)
			if !strings.HasPrefix(fDesc.Name, "Op") {
				continue
			}
			versions++
			if strings.HasSuffix(fDesc.Name,
				fmt.Sprintf("%d_%d", versionlib.MajorVersion, versionlib.MinorVersion)) {
				require.False(t, opV.Field(i).IsNil())
				nonemptyVersions++
			} else {
				require.True(t, opV.Field(i).IsNil())
			}
		}
		require.Equal(t, 1, nonemptyVersions)
		require.True(t, versions > 1)
	}
}
