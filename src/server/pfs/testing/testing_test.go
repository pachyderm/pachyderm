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
	"time"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	testSize = 200
)

func TestSimple(t *testing.T) {
	t.Parallel()
	apiClient := getPfsClient(t)
	repoName := uniqueString("testSimpleRepo")

	err := pfsclient.CreateRepo(apiClient, repoName)
	require.NoError(t, err)

	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.ID

	newCommitInfo, err := pfsclient.InspectCommit(apiClient, repoName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, newCommitInfo)
	require.Equal(t, newCommitID, newCommitInfo.Commit.ID)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_WRITE, newCommitInfo.CommitType)
	require.Nil(t, newCommitInfo.ParentCommit)

	commitInfos, err := pfsclient.ListCommit(apiClient, []string{repoName}, nil, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	require.Equal(t, newCommitInfo.Commit, commitInfos[0].Commit)

	err = pfsclient.MakeDirectory(apiClient, repoName, newCommitID, "a/b")
	require.NoError(t, err)
	err = pfsclient.MakeDirectory(apiClient, repoName, newCommitID, "a/c")
	require.NoError(t, err)
	err = pfsclient.MakeDirectory(apiClient, repoName, newCommitID, "a/d")
	require.NoError(t, err)

	doWrites(t, apiClient, repoName, newCommitID)

	err = pfsclient.FinishCommit(apiClient, repoName, newCommitID)
	require.NoError(t, err)

	newCommitInfo, err = pfsclient.InspectCommit(apiClient, repoName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, newCommitInfo)
	require.Equal(t, newCommitID, newCommitInfo.Commit.ID)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, newCommitInfo.CommitType)
	require.Nil(t, newCommitInfo.ParentCommit)

	checkWrites(t, apiClient, repoName, newCommitID)

	fileInfos, err := pfsclient.ListFile(
		apiClient,
		repoName,
		newCommitID,
		"a/b",
		"",
		&pfsclient.Shard{FileNumber: 0, BlockModulus: 1},
	)
	require.NoError(t, err)
	require.Equal(t, testSize, len(fileInfos))
	fileInfos, err = pfsclient.ListFile(
		apiClient,
		repoName,
		newCommitID,
		"a/c",
		"",
		&pfsclient.Shard{FileNumber: 0, BlockModulus: 1},
	)
	require.NoError(t, err)
	require.Equal(t, testSize, len(fileInfos))

	var fileInfos2 [7][]*pfsclient.FileInfo
	var wg sync.WaitGroup
	for i := 0; i < 7; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileInfos3, iErr := pfsclient.ListFile(
				apiClient,
				repoName,
				newCommitID,
				"a/b",
				"",
				&pfsclient.Shard{FileNumber: uint64(i), BlockModulus: 7},
			)
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

func TestBlockListCommits(t *testing.T) {
	t.Parallel()
	apiClient := getPfsClient(t)
	repoName := uniqueString("testBlockListCommitsRepo")

	err := pfsclient.CreateRepo(apiClient, repoName)
	require.NoError(t, err)

	baseCommit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	err = pfsclient.FinishCommit(apiClient, repoName, baseCommit.ID)
	require.NoError(t, err)

	repo := &pfsclient.Repo{
		Name: repoName,
	}
	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{repo},
		FromCommit: []*pfsclient.Commit{baseCommit},
	}
	commitInfos, err := apiClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	require.Equal(t, len(commitInfos.CommitInfo), 0)

	var newCommit *pfsclient.Commit
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1)
		commit, err := pfsclient.StartCommit(apiClient, repoName, baseCommit.ID, "")
		require.NoError(t, err)
		require.NotNil(t, commit)
		newCommit = commit
	}()
	listCommitRequest.Block = true
	listCommitRequest.CommitType = pfsclient.CommitType_COMMIT_TYPE_WRITE
	commitInfos, err = apiClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	wg.Wait()
	require.NoError(t, err)
	require.Equal(t, len(commitInfos.CommitInfo), 1)
	require.Equal(t, newCommit, commitInfos.CommitInfo[0].Commit)

	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1)
		err := pfsclient.FinishCommit(apiClient, repoName, newCommit.ID)
		require.NoError(t, err)
	}()
	listCommitRequest.Block = true
	listCommitRequest.CommitType = pfsclient.CommitType_COMMIT_TYPE_READ
	commitInfos, err = apiClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	wg.Wait()
	require.NoError(t, err)
	require.Equal(t, len(commitInfos.CommitInfo), 1)
	require.Equal(t, newCommit, commitInfos.CommitInfo[0].Commit)
}

func TestMount(t *testing.T) {
	t.Parallel()
	apiClient := getPfsClient(t)
	repoName := uniqueString("testMountRepo")

	err := pfsclient.CreateRepo(apiClient, repoName)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter("localhost", apiClient)
	ready := make(chan bool)
	go func() {
		err = mounter.Mount(directory, &pfsclient.Shard{FileNumber: 0, BlockModulus: 1}, nil, ready)
		require.NoError(t, err)
	}()
	<-ready

	_, err = os.Stat(filepath.Join(directory, repoName))
	require.NoError(t, err)

	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.ID

	_, err = os.Stat(filepath.Join(directory, repoName, newCommitID))
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(directory, repoName, newCommitID, "foo"), []byte("foo"), 0666)
	require.NoError(t, err)

	_, err = pfsclient.PutFile(apiClient, repoName, newCommitID, "bar", strings.NewReader("bar"))
	require.NoError(t, err)

	bigValue := make([]byte, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		bigValue[i] = 'a'
	}

	err = ioutil.WriteFile(filepath.Join(directory, repoName, newCommitID, "big1"), bigValue, 0666)
	require.NoError(t, err)

	_, err = pfsclient.PutFile(apiClient, repoName, newCommitID, "big2", bytes.NewReader(bigValue))
	require.NoError(t, err)

	err = pfsclient.FinishCommit(apiClient, repoName, newCommitID)
	require.NoError(t, err)

	fInfo, err := os.Stat(filepath.Join(directory, repoName, newCommitID, "foo"))
	require.NoError(t, err)
	require.Equal(t, int64(3), fInfo.Size())

	data, err := ioutil.ReadFile(filepath.Join(directory, repoName, newCommitID, "foo"))
	require.NoError(t, err)
	require.Equal(t, "foo", string(data))

	data, err = ioutil.ReadFile(filepath.Join(directory, repoName, newCommitID, "bar"))
	require.NoError(t, err)
	require.Equal(t, "bar", string(data))

	data, err = ioutil.ReadFile(filepath.Join(directory, repoName, newCommitID, "big1"))
	require.NoError(t, err)
	require.Equal(t, bigValue, data)

	data, err = ioutil.ReadFile(filepath.Join(directory, repoName, newCommitID, "big2"))
	require.NoError(t, err)
	require.Equal(t, bigValue, data)

	err = mounter.Unmount(directory)
	require.NoError(t, err)
}

func TestMountBig(t *testing.T) {
	t.Skip()
	t.Parallel()
	apiClient := getPfsClient(t)
	repoName := uniqueString("testMountBigRepo")

	err := pfsclient.CreateRepo(apiClient, repoName)
	require.NoError(t, err)

	directory := "/compile/testMount"
	mounter := fuse.NewMounter("localhost", apiClient)
	ready := make(chan bool)
	go func() {
		err = mounter.Mount(directory, &pfsclient.Shard{FileNumber: 0, BlockModulus: 1}, nil, ready)
		require.NoError(t, err)
	}()
	<-ready

	_, err = os.Stat(filepath.Join(directory, repoName))
	require.NoError(t, err)

	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	require.NotNil(t, commit)
	newCommitID := commit.ID

	bigValue := make([]byte, 1024*1024*300)
	for i := 0; i < 1024*1024*300; i++ {
		bigValue[i] = 'a'
	}

	wg := sync.WaitGroup{}
	for j := 0; j < 5; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			err := ioutil.WriteFile(filepath.Join(directory, repoName, newCommitID, fmt.Sprintf("big%d", j)), bigValue, 0666)
			require.NoError(t, err)
		}(j)
	}
	wg.Wait()

	err = pfsclient.FinishCommit(apiClient, repoName, newCommitID)
	require.NoError(t, err)

	wg = sync.WaitGroup{}
	for j := 0; j < 5; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			data, err := ioutil.ReadFile(filepath.Join(directory, repoName, newCommitID, fmt.Sprintf("big%d", j)))
			require.NoError(t, err)
			require.Equal(t, bigValue, data)
		}(j)
	}
	wg.Wait()

	err = mounter.Unmount(directory)
	require.NoError(t, err)
}

func BenchmarkFuse(b *testing.B) {
	apiClient := getPfsClient(b)
	repoName := uniqueString("benchMountRepo")

	if err := pfsclient.CreateRepo(apiClient, repoName); err != nil {
		b.Error(err)
	}

	directory := "/compile/benchMount"
	mounter := fuse.NewMounter("localhost", apiClient)
	ready := make(chan bool)
	go func() {
		err := mounter.Mount(directory, &pfsclient.Shard{FileNumber: 0, BlockModulus: 1}, nil, ready)
		require.NoError(b, err)
	}()
	<-ready

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
		commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
		if err != nil {
			b.Error(err)
		}
		if commit == nil {
			b.Error("nil branch")
		}
		newCommitID := commit.ID
		var wg sync.WaitGroup
		for j := 0; j < 1024; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				if err = ioutil.WriteFile(filepath.Join(directory, repoName, newCommitID, fmt.Sprintf("big%d", j)), bigValue, 0666); err != nil {
					b.Error(err)
				}
			}(j)
		}
		wg.Wait()
		if err := pfsclient.FinishCommit(apiClient, repoName, newCommitID); err != nil {
			b.Error(err)
		}
	}
}

func doWrites(tb testing.TB, apiClient pfsclient.APIClient, repoName string, commitID string) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, iErr := pfsclient.PutFile(apiClient, repoName, commitID,
				fmt.Sprintf("a/b/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(tb, iErr)
			_, iErr = pfsclient.PutFile(apiClient, repoName, commitID,
				fmt.Sprintf("a/c/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(tb, iErr)
		}()
	}
}

func checkWrites(tb testing.TB, apiClient pfsclient.APIClient, repoName string, commitID string) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := bytes.NewBuffer(nil)
			iErr := pfsclient.GetFile(
				apiClient,
				repoName,
				commitID,
				fmt.Sprintf("a/b/file%d", i),
				0,
				0,
				"",
				&pfsclient.Shard{FileNumber: 0, BlockModulus: 1},
				buffer,
			)
			require.NoError(tb, iErr)
			require.Equal(tb, fmt.Sprintf("hello%d", i), buffer.String())

			buffer = bytes.NewBuffer(nil)
			iErr = pfsclient.GetFile(
				apiClient,
				repoName,
				commitID,
				fmt.Sprintf("a/c/file%d", i),
				0,
				0,
				"",
				&pfsclient.Shard{FileNumber: 0, BlockModulus: 1},
				buffer,
			)
			require.NoError(tb, iErr)
			require.Equal(tb, fmt.Sprintf("hello%d", i), buffer.String())

		}()
	}
}

func getPfsClient(tb testing.TB) pfsclient.APIClient {
	pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if pfsdAddr == "" {
		if !testing.Short() {
			tb.Error("PFSD_PORT_650_TCP_ADDR not set")
		}
		tb.Skip("PFSD_PORT_650_TCP_ADDR not set")
	}
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:650", pfsdAddr), grpc.WithInsecure())
	require.NoError(tb, err)
	return pfsclient.NewAPIClient(clientConn)
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
