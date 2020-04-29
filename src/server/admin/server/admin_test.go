package server

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	versionlib "github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"

	"github.com/golang/snappy"
)

const (
	KB = 1024
	MB = 1024 * KB
)

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

func listAllCommits(t testing.TB, c *client.APIClient) []*pfs.CommitInfo {
	var commits []*pfs.CommitInfo
	repos, err := c.ListRepo()
	require.NoError(t, err)
	for _, r := range repos {
		cis, err := c.ListCommit(r.Repo.Name, "", "", 0)
		require.NoError(t, err)
		commits = append(commits, cis...)
	}
	return commits
}

// repoInfoSummary is a helper that converts a CommitINfo to a unique string
// representation, for checking the repos in a cluster before/after
// extract+restore.
func repoInfoSummary(i interface{}) interface{} {
	ri := i.(*pfs.RepoInfo)
	branchNames := make([]string, 0, len(ri.Branches))
	for _, b := range ri.Branches {
		branchNames = append(branchNames, b.Name)
	}
	sort.Strings(branchNames)
	return fmt.Sprintf("%s (%s) [Sz:%d] %v",
		ri.Repo.Name, ri.Description, ri.SizeBytes, branchNames)
}

// commitInfoSummary is a helper that converts a CommitInfo to a unique string
// representation, for checking the commits in a cluster before/after
// extract+restore.
// TODO(msteffen) This is similar to a lot of code in
// src/server/transaction/server/testing/server_test.go All of this should be
// unified in a library.
func commitInfoSummary(i interface{}) interface{} {
	ci := i.(*pfs.CommitInfo)
	parentCommit := "nil"
	if ci.ParentCommit != nil {
		parentCommit = ci.ParentCommit.ID
	}
	childCommits := func() []string {
		ccs := make([]string, 0, len(ci.ChildCommits))
		for _, c := range ci.ChildCommits {
			ccs = append(ccs, c.ID)
		}
		sort.Strings(ccs)
		return ccs
	}()
	prov := func() []string {
		ps := make([]string, 0, len(ci.Provenance))
		for _, p := range ci.Provenance {
			ps = append(ps, fmt.Sprintf("%s (%s)", p.Commit.ID, p.Branch.Name))
		}
		sort.Strings(ps)
		return ps
	}()
	subv := func() []string {
		ss := make([]string, 0, len(ci.Subvenance))
		for _, s := range ci.Subvenance {
			ss = append(ss, fmt.Sprintf("%s:%s", s.Lower.ID, s.Upper.ID))
		}
		sort.Strings(ss)
		return ss
	}()
	finished := "nil"
	if ci.Finished != nil {
		finished = types.TimestampString(ci.Finished)
	}
	tree := "nil"
	if ci.Tree != nil {
		tree = ci.Tree.Hash
	}
	trees := func() []string {
		ts := make([]string, 0, len(ci.Trees))
		for _, t := range ci.Trees {
			ts = append(ts, t.Hash)
		}
		sort.Strings(ts)
		return ts
	}()
	datums := "nil"
	if ci.Datums != nil {
		datums = ci.Datums.Hash
	}
	return fmt.Sprintf(`%s
Parent: %s
Children: %s
Provenance: %s
Subvenance: %s
Started: %s
Finished: %s
Size: %d
Tree: %s
Trees: %s
Datums: %s`,
		ci.Commit.ID,
		parentCommit, childCommits,
		prov, subv,
		types.TimestampString(ci.Started), finished,
		ci.SizeBytes, tree, trees, datums,
	)
}

func addData(t testing.TB, c *client.APIClient, repo string, count int, size uint) (fileHashes []string) {
	t.Helper()
	// Check for existing files
	var start int64
	if files, err := c.ListFile(repo, "master", "/"); err == nil {
		for _, fi := range files {
			dashI := strings.Index(fi.File.Path, "-")
			fileI, err := strconv.ParseInt(fi.File.Path[dashI+1:], 10, 64)
			require.NoError(t, err)
			if fileI > start {
				start = fileI
			}
		}
	}

	// Create new files
	r := rand.New(rand.NewSource(45 + start)) // write new data if addData called >1
	fileHashes = make([]string, 0, count)
	for i := 0; i < count; i++ {
		hash := md5.New()
		fileContent := workload.RandString(r, 40*MB)
		_, err := c.PutFile(repo, "master", fmt.Sprintf("file-%d", i),
			io.TeeReader(strings.NewReader(fileContent), hash))
		require.NoError(t, err, "error creating commits in %q: %v", repo, err)
		fileHashes = append(fileHashes, hex.EncodeToString(hash.Sum(nil)))
	}
	return fileHashes
}

// TODO(msteffen) equivalent to function in src/server/auth/server/admin_test.go.
// These should be unified.
func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

// testExtractRestore effectively implements both TestExtractRestoreObjects
// TestExtractRestoreNoObjects, as their logic is mostly the same
func testExtractRestore(t *testing.T, testObjects bool) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestoreObjects-in-")
	require.NoError(t, c.CreateRepo(dataRepo))

	// Create input data
	nCommits := 2
	fileHashes := addData(t, c, dataRepo, nCommits, 40*MB)

	// Create test pipelines
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
			&pps.ParallelismSpec{Constant: 1},
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
	// job's output commit closing and the job being marked successful, thus
	// retry)
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		jobInfos, err := c.ListJob("", nil, nil, -1, false)
		if err != nil {
			return err
		}
		// Pipelines were created after commits, so only the HEAD commits of the
		// input repo should be processed by each pipeline
		if len(jobInfos) != numPipelines {
			return errors.Errorf("expected %d jobs, but only encountered %d",
				numPipelines, len(jobInfos))
		}
		for _, ji := range jobInfos {
			if ji.State != pps.JobState_JOB_SUCCESS {
				return errors.Errorf("expected job %q to be in state SUCCESS but was %q",
					ji.Job.ID, ji.State.String())
			}
		}
		return nil
	})

	risBefore, err := c.ListRepo()
	require.NoError(t, err)
	commitInfosBefore := listAllCommits(t, c)
	// Extract existing cluster state
	ops, err := c.ExtractAll(testObjects)
	require.NoError(t, err)

	require.NoError(t, c.DeleteAll())

	if testObjects {
		// Delete existing objects
		require.NoError(t, c.GarbageCollect(10000))
	}

	// Restore metadata and possibly objects
	require.NoError(t, c.Restore(ops))
	// Do a fsck just in case.
	require.NoError(t, c.FsckFastExit())

	risAfter, err := c.ListRepo()
	require.NoError(t, err)
	require.ImagesEqual(t, risBefore, risAfter, repoInfoSummary)
	commitInfosAfter := listAllCommits(t, c)
	require.ImagesEqual(t, commitInfosBefore, commitInfosAfter, commitInfoSummary)

	// Make sure all commits got re-created
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		commitInfos, err := c.ListCommit(dataRepo, "", "", 0)
		if err != nil {
			return err
		}
		if len(commitInfos) != nCommits {
			return errors.Errorf("expected %d commits, but only encountered %d in %q",
				nCommits, len(commitInfos), dataRepo)
		}
		return nil
	})

	// Wait for re-created pipelines to process recreated input data
	commitIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	// Confirm all the recreated jobs passed
	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	// One job per pipeline, as above
	require.Equal(t, numPipelines, len(jis))
	for _, ji := range jis {
		// Job must ultimately succeed
		require.Equal(t, "JOB_SUCCESS", ji.State.String())
		require.Equal(t, int64(2), ji.DataTotal)
		require.NotNil(t, ji.Started)
		require.NotNil(t, ji.Finished)
	}

	// Make sure all branches were recreated
	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))

	// Check input data
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

	// Check that spec commits made it ok
	pis, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, numPipelines, len(pis))
	for _, pi := range pis {
		c.GetFile(pi.SpecCommit.Repo.Name, pi.SpecCommit.ID, ppsconsts.SpecFile, 0, 0, ioutil.Discard)
	}

	// make more commits
	addData(t, c, dataRepo, nCommits, 40*MB)

	// Wait for pipelines to process input data
	commitInfos, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, numPipelines, len(commitInfos))
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

func TestExtractRestoreFailedJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	// Create input repo
	dataRepo := tu.UniqueString("input-")
	require.NoError(t, c.CreateRepo(dataRepo))

	// Create test pipeline
	pipeline := tu.UniqueString("pipeline-")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			// First job/datum will fail (one file) second won't (two files)
			"set -ex", // datum fails if 'test' fails
			fmt.Sprintf("ls /pfs/%s/*", dataRepo),
			fmt.Sprintf("ls /pfs/%s/* | wc -l", dataRepo),
			fmt.Sprintf("test $( ls /pfs/%s/* | wc -l ) -gt 1", dataRepo),
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(dataRepo, "/"),
		"",
		false,
	))

	// Create input data
	nCommits := 2
	fileHashes := addData(t, c, dataRepo, nCommits, 40*MB)

	// Wait for pipelines to process input data
	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	// confirm that all jobs are in the expected state (there may be a short delay
	// between the job's output commit closing and the job being marked
	// successful, thus retry)
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		jobInfos, err := c.ListJob("", nil, nil, -1, false)
		if err != nil {
			return err
		}
		// Pipelines were created after commits, so only the HEAD commits of the
		// input repo should be processed by each pipeline
		if len(jobInfos) != nCommits {
			return errors.Errorf("expected %d jobs, but only encountered %d",
				nCommits, len(jobInfos))
		}
		stateCounts := make(map[pps.JobState]int)
		for _, ji := range jobInfos {
			stateCounts[ji.State]++
		}
		require.Equal(t, 2, len(stateCounts))
		// First job should succeed
		require.Equal(t, 1, stateCounts[pps.JobState_JOB_SUCCESS])
		// Second job should fail
		require.Equal(t, 1, stateCounts[pps.JobState_JOB_FAILURE])
		return nil
	})

	// Collect repos to compare after restore
	risBefore, err := c.ListRepo()
	require.NoError(t, err)
	commitInfosBefore := listAllCommits(t, c)

	// Extract existing cluster state
	ops, err := c.ExtractAll(true)
	require.NoError(t, err)

	// Delete metadata & data
	require.NoError(t, c.DeleteAll())
	// Delete existing objects
	require.NoError(t, c.GarbageCollect(10000))

	// Restore metadata and data
	require.NoError(t, c.Restore(ops))
	// Do a fsck just in case.
	require.NoError(t, c.FsckFastExit())

	risAfter, err := c.ListRepo()
	require.NoError(t, err)
	require.ImagesEqual(t, risBefore, risAfter, repoInfoSummary)
	commitInfosAfter := listAllCommits(t, c)
	require.ImagesEqual(t, commitInfosBefore, commitInfosAfter, commitInfoSummary)

	// Make sure all commits got re-created
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		commitInfos, err := c.ListCommit(dataRepo, "", "", 0)
		if err != nil {
			return err
		}
		if len(commitInfos) != nCommits {
			return errors.Errorf("expected %d commits, but only encountered %d in %q",
				nCommits, len(commitInfos), dataRepo)
		}
		return nil
	})

	// Wait for re-created pipelines to process recreated input data
	commitIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	// Make sure all branches were recreated
	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))

	// Check input data
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

	// Check that spec commits were all restored
	pis, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 1, len(pis))
	for _, pi := range pis {
		c.GetFile(pi.SpecCommit.Repo.Name, pi.SpecCommit.ID, ppsconsts.SpecFile, 0, 0, ioutil.Discard)
	}

	// make more commits & wait for pipelines to process it (check cluster health)
	addData(t, c, dataRepo, nCommits, 40*MB)
	commitInfos, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestExtractRestoreHeadlessBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create a headless branch
	require.NoError(t, c.CreateBranch(dataRepo, "headless", "", nil))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))
	// Do a fsck just in case.
	require.NoError(t, c.FsckFastExit())

	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))
	require.Equal(t, "headless", bis[0].Branch.Name)
}

func TestExtractVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
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
		&pps.ParallelismSpec{Constant: 1},
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
	c := tu.GetPachClient(t)
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

func TestExtractRestorePipelineUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	input1 := tu.UniqueString("TestExtractRestorePipelineUpdate_data")
	require.NoError(t, c.CreateRepo(input1))

	pipeline := tu.UniqueString("TestExtractRestorePipelineUpdate")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input1),
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(input1, "/*"),
		"",
		false,
	))
	_, err := c.PutFile(input1, "master", "file", strings.NewReader("file"))
	require.NoError(t, err)
	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(input1, "master")}, nil)
	require.NoError(t, err)

	input2 := tu.UniqueString("TestExtractRestorePipelineUpdate_data")
	require.NoError(t, c.CreateRepo(input2))

	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input2),
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(input2, "/*"),
		"",
		true,
	))

	_, err = c.PutFile(input2, "master", "file", strings.NewReader("file"))
	require.NoError(t, err)
	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(input2, "master")}, nil)
	require.NoError(t, err)

	commitInfosBefore := listAllCommits(t, c)
	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))
	// Do a fsck just in case.
	require.NoError(t, c.FsckFastExit())

	// Check that commits still have the right provenance
	commitInfosAfter := listAllCommits(t, c)
	require.ImagesEqual(t, commitInfosBefore, commitInfosAfter, commitInfoSummary)
}

func TestExtractRestoreDeferredProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestoreDeferredProcessing_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline1 := tu.UniqueString("TestExtractRestoreDeferredProcessing")
	_, err := c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline1),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo)},
			},
			Input:        client.NewPFSInput(dataRepo, "/*"),
			OutputBranch: "staging",
		})
	require.NoError(t, err)

	pipeline2 := tu.UniqueString("TestExtractRestoreDeferredProcessing2")
	require.NoError(t, c.CreatePipeline(
		pipeline2,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipeline1),
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pipeline1, "/*"),
		"",
		false,
	))

	_, err = c.PutFile(dataRepo, "staging", "file", strings.NewReader("file"))
	require.NoError(t, err)
	c.CreateBranch(dataRepo, "master", "staging", nil)
	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)

	c.CreateBranch(pipeline1, "master", "staging", nil)

	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)

	commitInfosBefore := listAllCommits(t, c)

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))
	// Do a fsck just in case.
	require.NoError(t, c.FsckFastExit())

	// Check that commits are all the same after restore
	commitInfosAfter := listAllCommits(t, c)
	require.ImagesEqual(t, commitInfosBefore, commitInfosAfter, commitInfoSummary)

	_, err = c.PutFile(dataRepo, "staging", "file2", strings.NewReader("file"))
	require.NoError(t, err)
	c.CreateBranch(dataRepo, "master", "staging", nil)
	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)

	c.CreateBranch(pipeline1, "master", "staging", nil)

	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
}

func TestExtractRestoreStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestoreStats_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	_, err := c.PutFile(dataRepo, "master", "file", strings.NewReader("file"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("TestExtractRestoreStats")
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"echo foo",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewPFSInput(dataRepo, "/*"),
			EnableStats: true,
		})
	require.NoError(t, err)

	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)

	commitInfosBefore := listAllCommits(t, c)

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))

	// Check that commits are all the same after restore
	commitInfosAfter := listAllCommits(t, c)
	require.ImagesEqual(t, commitInfosBefore, commitInfosAfter, commitInfoSummary)

	cis, err := c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(cis))
}
