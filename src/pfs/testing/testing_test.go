package testing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"go.pedge.io/protolog/logrus"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
)

const (
	testSize = 100
)

func init() {
	logrus.Register()
}

func TestSimple(t *testing.T) {
	t.Parallel()
	RunTest(t, testSimple)
}

func TestFailures(t *testing.T) {
	t.Skip()
	t.Parallel()
	RunTest(t, testFailures)
}

func TestFuseMount(t *testing.T) {
	t.Skip()
	t.Parallel()
	RunTest(t, testMount)
}

func TestFuseMountBig(t *testing.T) {
	t.Skip()
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	RunTest(t, testMountBig)

}

func BenchmarkFuse(b *testing.B) {
	RunBench(b, benchMount)
}

func testSimple(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient, cluster Cluster) {
	repositoryName := TestRepositoryName()

	err := pfsutil.RepoCreate(apiClient, repositoryName)
	require.NoError(t, err)

	commitInfo, err := pfsutil.CommitInspect(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, commitInfo)
	require.Equal(t, "scratch", commitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Nil(t, commitInfo.ParentCommit)
	scratchCommitInfo := commitInfo

	commitInfos, err := pfsutil.CommitList(apiClient, repositoryName)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	require.Equal(t, scratchCommitInfo, commitInfos[0])

	commit, err := pfsutil.CommitStart(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.Id

	commitInfo, err = pfsutil.CommitInspect(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, commitInfo)
	require.Equal(t, newCommitID, commitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, "scratch", commitInfo.ParentCommit.Id)
	newCommitInfo := commitInfo

	commitInfos, err = pfsutil.CommitList(apiClient, repositoryName)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	require.Equal(t, newCommitInfo, commitInfos[0])
	require.Equal(t, scratchCommitInfo, commitInfos[1])

	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/b")
	require.NoError(t, err)
	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/c")
	require.NoError(t, err)

	doWrites(t, apiClient, repositoryName, newCommitID)

	err = pfsutil.CommitFinish(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	commitInfo, err = pfsutil.CommitInspect(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, commitInfo)
	require.Equal(t, newCommitID, commitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, "scratch", commitInfo.ParentCommit.Id)

	checkWrites(t, apiClient, repositoryName, newCommitID)

	fileInfos, err := pfsutil.FileList(apiClient, repositoryName, newCommitID, "a/b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(fileInfos))
	fileInfos, err = pfsutil.FileList(apiClient, repositoryName, newCommitID, "a/c", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(fileInfos))

	var fileInfos2 [7][]*pfs.FileInfo
	var wg sync.WaitGroup
	for i := 0; i < 7; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileInfos3, iErr := pfsutil.FileList(apiClient, repositoryName, newCommitID, "a/b", uint64(i), 7)
			require.NoError(t, iErr)
			fileInfos2[i] = fileInfos3
		}()
	}
	wg.Wait()
	count := 0
	for i := 0; i < 7; i++ {
		count += len(fileInfos2[i])
	}
	require.Equal(t, testSize, count)
}

func testFailures(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient, cluster Cluster) {
	repositoryName := TestRepositoryName()

	err := pfsutil.RepoCreate(apiClient, repositoryName)
	require.NoError(t, err)

	commit, err := pfsutil.CommitStart(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.Id

	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/b")
	require.NoError(t, err)
	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/c")
	require.NoError(t, err)

	doWrites(t, apiClient, repositoryName, newCommitID)

	err = pfsutil.CommitFinish(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	checkWrites(t, apiClient, repositoryName, newCommitID)

	for server := 0; server < testNumReplicas; server++ {
		cluster.Kill(server)
	}
	cluster.WaitForAvailability()

	checkWrites(t, apiClient, repositoryName, newCommitID)
}

func testMount(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient, cluster Cluster) {
	repositoryName := TestRepositoryName()

	err := pfsutil.RepoCreate(apiClient, repositoryName)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter(apiClient)
	err = mounter.Mount(repositoryName, directory, "", 0, 1)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(directory, "scratch"))
	require.NoError(t, err)

	commit, err := pfsutil.CommitStart(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.Id

	_, err = os.Stat(filepath.Join(directory, newCommitID))
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(directory, newCommitID, "foo"), []byte("foo"), 0666)
	require.NoError(t, err)

	_, err = pfsutil.FilePut(apiClient, repositoryName, newCommitID, "bar", 0, strings.NewReader("bar"))
	require.NoError(t, err)

	bigValue := make([]byte, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		bigValue[i] = 'a'
	}

	err = ioutil.WriteFile(filepath.Join(directory, newCommitID, "big1"), bigValue, 0666)
	require.NoError(t, err)

	_, err = pfsutil.FilePut(apiClient, repositoryName, newCommitID, "big2", 0, bytes.NewReader(bigValue))
	require.NoError(t, err)

	err = pfsutil.CommitFinish(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	fInfo, err := os.Stat(filepath.Join(directory, newCommitID, "foo"))
	require.NoError(t, err)
	require.Equal(t, int64(3), fInfo.Size())

	data, err := ioutil.ReadFile(filepath.Join(directory, newCommitID, "foo"))
	require.NoError(t, err)
	require.Equal(t, "foo", string(data))

	data, err = ioutil.ReadFile(filepath.Join(directory, newCommitID, "bar"))
	require.NoError(t, err)
	require.Equal(t, "bar", string(data))

	data, err = ioutil.ReadFile(filepath.Join(directory, newCommitID, "big1"))
	require.NoError(t, err)
	require.Equal(t, bigValue, data)

	data, err = ioutil.ReadFile(filepath.Join(directory, newCommitID, "big2"))
	require.NoError(t, err)
	require.Equal(t, bigValue, data)

	err = mounter.Unmount(directory)
	require.NoError(t, err)
	err = mounter.Wait(directory)
	require.NoError(t, err)
}

func testMountBig(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient, cluster Cluster) {
	repositoryName := TestRepositoryName()

	err := pfsutil.RepoCreate(apiClient, repositoryName)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter(apiClient)
	err = mounter.Mount(repositoryName, "", directory, 0, 1)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(directory, "scratch"))
	require.NoError(t, err)

	commit, err := pfsutil.CommitStart(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.Id

	bigValue := make([]byte, 1024*1024*300)
	for i := 0; i < 1024*1024*300; i++ {
		bigValue[i] = 'a'
	}

	wg := sync.WaitGroup{}
	for j := 0; j < 5; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			err := ioutil.WriteFile(filepath.Join(directory, newCommitID, fmt.Sprintf("big%d", j)), bigValue, 0666)
			require.NoError(t, err)
		}(j)
	}
	wg.Wait()

	err = pfsutil.CommitFinish(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	wg = sync.WaitGroup{}
	for j := 0; j < 5; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			data, err := ioutil.ReadFile(filepath.Join(directory, newCommitID, fmt.Sprintf("big%d", j)))
			require.NoError(t, err)
			require.Equal(t, bigValue, data)
		}(j)
	}
	wg.Wait()

	err = mounter.Unmount(directory)
	require.NoError(t, err)
	err = mounter.Wait(directory)
	require.NoError(t, err)
}

func benchMount(b *testing.B, apiClient pfs.ApiClient) {
	repositoryName := TestRepositoryName()

	if err := pfsutil.RepoCreate(apiClient, repositoryName); err != nil {
		b.Error(err)
	}

	directory := "/compile/benchMount"
	mounter := fuse.NewMounter(apiClient)
	if err := mounter.Mount(repositoryName, "", directory, 0, 1); err != nil {
		b.Error(err)
	}

	defer func() {
		if err := mounter.Unmount(directory); err != nil {
			b.Error(err)
		}
		if err := mounter.Wait(directory); err != nil {
			b.Error(err)
		}
	}()

	bigValue := make([]byte, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		bigValue[i] = 'a'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		commit, err := pfsutil.CommitStart(apiClient, repositoryName, "scratch")
		if err != nil {
			b.Error(err)
		}
		if commit == nil {
			b.Error("nil branch")
		}
		newCommitID := commit.Id
		var wg sync.WaitGroup
		for j := 0; j < 1024; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				if err = ioutil.WriteFile(filepath.Join(directory, newCommitID, fmt.Sprintf("big%d", j)), bigValue, 0666); err != nil {
					b.Error(err)
				}
			}(j)
		}
		wg.Wait()
		if err := pfsutil.CommitFinish(apiClient, repositoryName, newCommitID); err != nil {
			b.Error(err)
		}
	}
}

func doWrites(tb testing.TB, apiClient pfs.ApiClient, repositoryName string, commitID string) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, iErr := pfsutil.FilePut(apiClient, repositoryName, commitID,
				fmt.Sprintf("a/b/file%d", i), 0, strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(tb, iErr)
			_, iErr = pfsutil.FilePut(apiClient, repositoryName, commitID,
				fmt.Sprintf("a/c/file%d", i), 0, strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(tb, iErr)
		}()
	}
}

func checkWrites(tb testing.TB, apiClient pfs.ApiClient, repositoryName string, commitID string) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := bytes.NewBuffer(nil)
			iErr := pfsutil.FileGet(apiClient, repositoryName, commitID,
				fmt.Sprintf("a/b/file%d", i), 0, pfsutil.GetAll, buffer)
			require.NoError(tb, iErr)
			require.Equal(tb, fmt.Sprintf("hello%d", i), buffer.String())
			buffer = bytes.NewBuffer(nil)
			iErr = pfsutil.FileGet(apiClient, repositoryName, commitID,
				fmt.Sprintf("a/c/file%d", i), 0, pfsutil.GetAll, buffer)
			require.NoError(tb, iErr)
			require.Equal(tb, fmt.Sprintf("hello%d", i), buffer.String())
		}()
	}
}
