package server

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	versionlib "github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"

	"github.com/golang/snappy"
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
			pachClient, err = client.NewForTest()
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

// TODO(msteffen) equivalent to funciton in src/server/auth/server/admin_test.go.
// These should be unified.
func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

// testExtractRestored effectively implements both TestExtractRestoreObjects
// TestExtractRestoreNoObjects, as their logic is mostly the same
func testExtractRestore(t *testing.T, testObjects bool) {
	fmt.Printf(">>> TestExtractRestore(objects=%t)\n", testObjects)
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestoreObjects-in-")
	require.NoError(t, c.CreateRepo(dataRepo))

	// Create input data
	fmt.Printf(">>> Create input data\n")
	nCommits := 2
	r := rand.New(rand.NewSource(45))
	fileHashes := make([]string, 0, nCommits)
	for i := 0; i < nCommits; i++ {
		hash := md5.New()
		fileContent := workload.RandString(r, 40*MB)
		_, err := c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i),
			io.TeeReader(strings.NewReader(fileContent), hash))
		require.NoError(t, err)
		fileHashes = append(fileHashes, hex.EncodeToString(hash.Sum(nil)))
	}

	// Create test pipelines
	fmt.Printf(">>> Create test pipelines\n")
	numPipelines := 3
	var input, pipeline string
	input = dataRepo
	for i := 0; i < numPipelines; i++ {
		pipeline = tu.UniqueString(fmt.Sprintf("TestExtractRestoreObjects-P%d-", i))
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input),
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

	// Wait for pipelines to process input data
	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	// confirm that all the jobs passed (there may be a short delay between the
	// job's output commit closing and the job being marked successful, thus retry
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		jobInfos, err := c.ListJob("", nil, nil, -1, false)
		if err != nil {
			return err
		}
		// Pipelines were created after commits, so only the HEAD commits of the
		// input repo should be processed by each pipeline
		if len(jobInfos) != numPipelines {
			return fmt.Errorf("expected %d commits, but only encountered %d",
				nCommits*numPipelines, len(jobInfos))
		}
		for _, ji := range jobInfos {
			if ji.State != pps.JobState_JOB_SUCCESS {
				return fmt.Errorf("expected job %q to be in state SUCCESS but was %q",
					ji.Job.ID, ji.State.String())
			}
		}
		return nil
	})

	// Extract existing cluster state
	fmt.Printf(">>> Extract existing cluster state\n")
	ops, err := c.ExtractAll(testObjects)
	require.NoError(t, err)

	// Delete existing metadata
	fmt.Printf(">>> Delete existing metadata\n")
	require.NoError(t, c.DeleteAll())

	if testObjects {
		// Delete existing objects
		fmt.Printf(">>> Delete existing objects\n")
		require.NoError(t, c.GarbageCollect(10000))
	}

	// Restore metadata and possibly objects
	fmt.Printf(">>> Restore metadata and possibly objects\n")
	require.NoError(t, c.Restore(ops))

	// Make sure all commits got re-created
	fmt.Printf(">>> Make sure all commits got re-created\n")
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		commitInfos, err := c.ListCommit(dataRepo, "", "", 0)
		if err != nil {
			return err
		}
		if len(commitInfos) != nCommits {
			return fmt.Errorf("expected %d commits, but only encountered %d in %q",
				nCommits, len(commitInfos), dataRepo)
		}
		return nil
	})

	// Wait for re-created pipelines to process recreated input data
	fmt.Printf(">>> Wait for re-created pipelines to process recreated input data\n")
	commitIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	// Confirm all the recreated jobs passed
	fmt.Printf(">>> Confirm all the recreated jobs passed\n")
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		fmt.Printf(">>> Calling ListJob\n")
		jobInfos, err := c.ListJob("", nil, nil, -1, false) // make sure jobs all succeeded
		if err != nil {
			return err
		}
		// Extract places commits before pipelines, so that all commits are processed
		// after Restore() (thus |jobInfos| = nCommits * numPipelines)
		if len(jobInfos) != nCommits*numPipelines {
			return fmt.Errorf("expected %d commits, but only encountered %d",
				nCommits*numPipelines, len(jobInfos))
		}
		for _, ji := range jobInfos {
			// race--we may call listJob between when a job's output commit is closed
			// and when its state is updated
			if ji.State != pps.JobState_JOB_SUCCESS {
				return fmt.Errorf("expected job %q to be in state JOB_SUCCESS but was %q",
					ji.Job.ID, ji.State.String())
			}
		}
		return nil
	})

	// Make sure all branches were recreated
	fmt.Printf(">>> Make sure all branches were recreated\n")
	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))

	// Check input data
	fmt.Printf(">>> Check input data\n")
	// This check uses a backoff because sometimes GetFile causes pachd to OOM
	var restoredFileHashes []string
	require.NoError(t, backoff.Retry(func() error {
		restoredFileHashes = make([]string, 0, nCommits)
		for i := 0; i < nCommits; i++ {
			hash := md5.New()
			err := c.GetFile(dataRepo, "master", fmt.Sprintf("file-%d", i), 0, 0, hash)
			if err != nil {
				return err
			}
			restoredFileHashes = append(restoredFileHashes, hex.EncodeToString(hash.Sum(nil)))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	require.ElementsEqual(t, fileHashes, restoredFileHashes)

	// Check output data
	fmt.Printf(">>> Check output data\n")
	// This check uses a backoff because sometimes GetFile causes pachd to OOM
	require.NoError(t, backoff.Retry(func() error {
		restoredFileHashes = make([]string, 0, nCommits)
		for i := 0; i < nCommits; i++ {
			hash := md5.New()
			err := c.GetFile(pipeline, "master", fmt.Sprintf("file-%d", i), 0, 0, hash)
			if err != nil {
				return err
			}
			restoredFileHashes = append(restoredFileHashes, hex.EncodeToString(hash.Sum(nil)))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	require.ElementsEqual(t, fileHashes, restoredFileHashes)
}

// TestExtractRestoreNoObjects tests extraction and restoration in the case
// where existing objects are re-used (common for cloud deployments, as objects
// are stored outside of kubernetes, in object store)
func TestExtractRestoreNoObjects(t *testing.T) {
	testExtractRestore(t, false)
}

// TestExtractRestoreObjects tests extraction and restoration of objects. Note
// that since 1.8, only data in input repos is referenced by objects, so this
// tests extracting/restoring an input repo.
func TestExtractRestoreObjects(t *testing.T) {
	testExtractRestore(t, true)
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

func TestMigrateFrom1_7(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Clear pachyderm cluster (so that next cluster starts up in a clean environment)
	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	// Restore dumped metadata (now that objects are present)
	md, err := os.Open(path.Join(os.Getenv("GOPATH"),
		"src/github.com/pachyderm/pachyderm/etc/testing/migration/v1_7/sort.metadata"))
	require.NoError(t, err)
	require.NoError(t, c.RestoreReader(snappy.NewReader(md)))
	require.NoError(t, md.Close())

	// Wait for final imported commit to be processed
	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit("left", "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	// filter-left and filter-right both compute a join of left and
	// right--depending on when the final commit to 'left' was added, it may have
	// been processed multiple times (should be n * 3, as there are 3 pipelines)
	require.True(t, len(commitInfos) >= 3)

	// Inspect input
	commits, err := c.ListCommit("left", "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commits))
	commits, err = c.ListCommit("right", "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commits))

	// Inspect output
	repos, err := c.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t,
		[]string{"left", "right", "copy", "sort"},
		repos, RepoInfoToName)

	// make sure all numbers 0-99 are in /nums
	var buf bytes.Buffer
	require.NoError(t, c.GetFile("sort", "master", "/nums", 0, 0, &buf))
	s := bufio.NewScanner(&buf)
	numbers := make(map[string]struct{})
	for s.Scan() {
		numbers[s.Text()] = struct{}{}
	}
	require.Equal(t, 100, len(numbers)) // job processed all inputs

	// Confirm stats commits are present
	commits, err = c.ListCommit("sort", "stats", "", 0)
	require.NoError(t, err)
	require.Equal(t, 6, len(commits))
}

func int64p(i int64) *int64 {
	return &i
}
