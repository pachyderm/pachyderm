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
	"github.com/stretchr/testify/require"
)

const (
	testSize = 100
)

func init() {
	logrus.Register()
}

func TestBtrfs(t *testing.T) {
	t.Parallel()
	RunTest(t, testSimple)
}

func TestReplica(t *testing.T) {
	t.Parallel()
	RunTest(t, testReplica)
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

func testSimple(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient) {
	repositoryName := TestRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName, false)
	require.NoError(t, err)

	getCommitInfoResponse, err := pfsutil.GetCommitInfo(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Nil(t, getCommitInfoResponse.CommitInfo.ParentCommit)

	branchResponse, err := pfsutil.Branch(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, branchResponse)
	newCommitID := branchResponse.Commit.Id

	getCommitInfoResponse, err = pfsutil.GetCommitInfo(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_WRITE, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)

	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/b")
	require.NoError(t, err)
	err = pfsutil.MakeDirectory(apiClient, repositoryName, newCommitID, "a/c")
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, iErr := pfsutil.PutFile(apiClient, repositoryName, newCommitID,
				fmt.Sprintf("a/b/file%d", i), 0, strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
			_, iErr = pfsutil.PutFile(apiClient, repositoryName, newCommitID,
				fmt.Sprintf("a/c/file%d", i), 0, strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
		}()
	}
	wg.Wait()

	err = pfsutil.Commit(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	getCommitInfoResponse, err = pfsutil.GetCommitInfo(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)

	wg = sync.WaitGroup{}
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := bytes.NewBuffer(nil)
			iErr := pfsutil.GetFile(apiClient, repositoryName, newCommitID,
				fmt.Sprintf("a/b/file%d", i), 0, pfsutil.GetAll, buffer)
			require.NoError(t, iErr)
			require.Equal(t, fmt.Sprintf("hello%d", i), buffer.String())
			buffer = bytes.NewBuffer(nil)
			iErr = pfsutil.GetFile(apiClient, repositoryName, newCommitID,
				fmt.Sprintf("a/c/file%d", i), 0, pfsutil.GetAll, buffer)
			require.NoError(t, iErr)
			require.Equal(t, fmt.Sprintf("hello%d", i), buffer.String())
		}()
	}
	wg.Wait()

	listFilesResponse, err := pfsutil.ListFiles(apiClient, repositoryName, newCommitID, "a/b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))
	listFilesResponse, err = pfsutil.ListFiles(apiClient, repositoryName, newCommitID, "a/c", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))

	var fileInfos [7][]*pfs.FileInfo
	wg = sync.WaitGroup{}
	for i := 0; i < 7; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			listFilesResponse, iErr := pfsutil.ListFiles(apiClient, repositoryName, newCommitID, "a/b", uint64(i), 7)
			require.NoError(t, iErr)
			fileInfos[i] = listFilesResponse.FileInfo
		}()
	}
	wg.Wait()
	count := 0
	for i := 0; i < 7; i++ {
		count += len(fileInfos[i])
	}
	require.Equal(t, testSize, count)
}

func testReplica(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient) {
	masterName := TestRepositoryName()
	err := pfsutil.InitRepository(apiClient, masterName, false)
	require.NoError(t, err)

	replicaName := TestRepositoryName()
	err = pfsutil.InitRepository(apiClient, replicaName, true)
	require.NoError(t, err)

	for i := 0; i < testShardsPerServer*testNumServers; i++ {
		var buffer bytes.Buffer
		err := pfsutil.PullDiff(internalAPIClient, masterName, "scratch", uint64(i), &buffer)
		require.NoError(t, err)
		err = pfsutil.PushDiff(internalAPIClient, replicaName, "scratch", uint64(i), &buffer)
		require.NoError(t, err)
	}

	getCommitInfoResponse, err := pfsutil.GetCommitInfo(apiClient, replicaName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.NotNil(t, getCommitInfoResponse.CommitInfo)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Nil(t, getCommitInfoResponse.CommitInfo.ParentCommit)

	branchResponse, err := pfsutil.Branch(apiClient, masterName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, branchResponse)
	newCommitID := branchResponse.Commit.Id

	var wg sync.WaitGroup
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, iErr := pfsutil.PutFile(apiClient, masterName, newCommitID,
				fmt.Sprintf("file%d", i), 0, strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
		}()
	}
	wg.Wait()

	err = pfsutil.Commit(apiClient, masterName, newCommitID)
	require.NoError(t, err)

	for i := 0; i < testShardsPerServer*testNumServers; i++ {
		var buffer bytes.Buffer
		err := pfsutil.PullDiff(internalAPIClient, masterName, newCommitID, uint64(i), &buffer)
		require.NoError(t, err)
		err = pfsutil.PushDiff(internalAPIClient, replicaName, newCommitID, uint64(i), &buffer)
		require.NoError(t, err)
	}

	getCommitInfoResponse, err = pfsutil.GetCommitInfo(apiClient, replicaName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.NotNil(t, getCommitInfoResponse.CommitInfo)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)
}

func testMount(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient) {
	repositoryName := TestRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName, false)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter()
	go func() {
		err = mounter.Mount(apiClient, repositoryName, directory, 0, 1)
		require.NoError(t, err)
	}()
	mounter.Ready()

	_, err = os.Stat(filepath.Join(directory, "scratch"))
	require.NoError(t, err)

	branchResponse, err := pfsutil.Branch(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, branchResponse)
	newCommitID := branchResponse.Commit.Id

	_, err = os.Stat(filepath.Join(directory, newCommitID))
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(directory, newCommitID, "foo"), []byte("foo"), 0666)
	require.NoError(t, err)

	_, err = pfsutil.PutFile(apiClient, repositoryName, newCommitID, "bar", 0, strings.NewReader("bar"))
	require.NoError(t, err)

	bigValue := make([]byte, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		bigValue[i] = 'a'
	}

	err = ioutil.WriteFile(filepath.Join(directory, newCommitID, "big1"), bigValue, 0666)
	require.NoError(t, err)

	_, err = pfsutil.PutFile(apiClient, repositoryName, newCommitID, "big2", 0, bytes.NewReader(bigValue))
	require.NoError(t, err)

	err = pfsutil.Commit(apiClient, repositoryName, newCommitID)
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
}

func testMountBig(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient) {
	repositoryName := TestRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName, false)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter()
	go func() {
		err = mounter.Mount(apiClient, repositoryName, directory, 0, 1)
		require.NoError(t, err)
	}()
	mounter.Ready()

	_, err = os.Stat(filepath.Join(directory, "scratch"))
	require.NoError(t, err)

	branchResponse, err := pfsutil.Branch(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, branchResponse)
	newCommitID := branchResponse.Commit.Id

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

	err = pfsutil.Commit(apiClient, repositoryName, newCommitID)
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
}

func benchMount(b *testing.B, apiClient pfs.ApiClient) {
	repositoryName := TestRepositoryName()

	if err := pfsutil.InitRepository(apiClient, repositoryName, false); err != nil {
		b.Error(err)
	}

	directory := "/compile/benchMount"
	mounter := fuse.NewMounter()
	go func() {
		if err := mounter.Mount(apiClient, repositoryName, directory, 0, 1); err != nil {
			b.Error(err)
		}
	}()
	mounter.Ready()

	defer func() {
		if err := mounter.Unmount(directory); err != nil {
			b.Error(err)
		}
	}()

	bigValue := make([]byte, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		bigValue[i] = 'a'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		branchResponse, err := pfsutil.Branch(apiClient, repositoryName, "scratch")
		if err != nil {
			b.Error(err)
		}
		if branchResponse == nil {
			b.Error("nil branch")
		}
		newCommitID := branchResponse.Commit.Id
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
		if err := pfsutil.Commit(apiClient, repositoryName, newCommitID); err != nil {
			b.Error(err)
		}
	}
}
