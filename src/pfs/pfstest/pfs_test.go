package pfstest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/pachyderm/pachyderm/src/pkg/grpctest"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/stretchr/testify/require"
)

const (
	// TODO(pedge): large numbers of shards takes forever because
	// we are doing tons of btrfs operations on init, is there anything
	// we can do about that?
	testShardsPerServer = 4
	testNumServers      = 1
	testSize            = 10000
)

var (
	counter int32
)

func init() {
	executil.SetDebug(true)
}

func TestBtrfs(t *testing.T) {
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	runTest(t, driver, testSimple)
}

func getBtrfsRootDir(t *testing.T) string {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_BTRFS_ROOT")
	if rootDir == "" {
		t.Fatal("PFS_BTRFS_ROOT not set")
	}
	return rootDir
}

func testSimple(t *testing.T, apiClient pfs.ApiClient) {
	repositoryName := testRepositoryName()

	err := initRepository(apiClient, repositoryName)
	require.NoError(t, err)

	getCommitInfoResponse, err := getCommitInfo(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Nil(t, getCommitInfoResponse.CommitInfo.ParentCommit)

	branchResponse, err := branch(apiClient, repositoryName, "scratch")
	require.NoError(t, err)
	require.NotNil(t, branchResponse)
	newCommitID := branchResponse.Commit.Id

	getCommitInfoResponse, err = getCommitInfo(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_WRITE, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)

	err = makeDirectory(apiClient, repositoryName, newCommitID, "a/b")
	require.NoError(t, err)
	err = makeDirectory(apiClient, repositoryName, newCommitID, "a/c")
	require.NoError(t, err)

	for i := 0; i < testSize; i++ {
		err = putFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
		require.NoError(t, err)
		err = putFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
		require.NoError(t, err)
	}

	err = commit(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	getCommitInfoResponse, err = getCommitInfo(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)

	for i := 0; i < testSize; i++ {
		readStringer, err := getFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i))
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("hello%d", i), readStringer.String())
		readStringer, err = getFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i))
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("hello%d", i), readStringer.String())
	}

	listFilesResponse, err := listFiles(apiClient, repositoryName, newCommitID, "a/b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))
	listFilesResponse, err = listFiles(apiClient, repositoryName, newCommitID, "a/c", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))
	listFilesResponse, err = listFiles(apiClient, repositoryName, newCommitID, "a/b", 0, 2)
	require.NoError(t, err)
	require.Equal(t, testSize/2, len(listFilesResponse.FileInfo))
	listFilesResponse, err = listFiles(apiClient, repositoryName, newCommitID, "a/c", 1, 2)
	require.NoError(t, err)
	require.Equal(t, testSize/2, len(listFilesResponse.FileInfo))
}

func testRepositoryName() string {
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}

func initRepository(apiClient pfs.ApiClient, repositoryName string) error {
	_, err := apiClient.InitRepository(
		context.Background(),
		&pfs.InitRepositoryRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
	return err
}

func branch(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.BranchResponse, error) {
	return apiClient.Branch(
		context.Background(),
		&pfs.BranchRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

func makeDirectory(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) error {
	_, err := apiClient.MakeDirectory(
		context.Background(),
		&pfs.MakeDirectoryRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	return err
}

func putFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, reader io.Reader) error {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Value: value,
		},
	)
	return err
}

type readStringer interface {
	io.Reader
	fmt.Stringer
}

func getFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) (readStringer, error) {
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	if err := protoutil.WriteFromStreamingBytesClient(apiGetFileClient, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

func listFiles(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, shardNum int, shardModulo int) (*pfs.ListFilesResponse, error) {
	return apiClient.ListFiles(
		context.Background(),
		&pfs.ListFilesRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Shard: &pfs.Shard{
				Number: uint64(shardNum),
				Modulo: uint64(shardModulo),
			},
		},
	)
}

func commit(apiClient pfs.ApiClient, repositoryName string, commitID string) error {
	_, err := apiClient.Commit(
		context.Background(),
		&pfs.CommitRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
	return err
}

func getCommitInfo(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.GetCommitInfoResponse, error) {
	return apiClient.GetCommitInfo(
		context.Background(),
		&pfs.GetCommitInfoRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

func runTest(
	t *testing.T,
	driver drive.Driver,
	f func(t *testing.T, apiClient pfs.ApiClient),
) {
	grpctest.Run(
		t,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			for address, s := range servers {
				combinedAPIServer := server.NewCombinedAPIServer(
					route.NewSharder(
						testShardsPerServer*testNumServers,
					),
					route.NewRouter(
						route.NewSingleAddresser(
							address,
							testShardsPerServer,
						),
						route.NewDialer(),
						address,
					),
					driver,
				)
				pfs.RegisterApiServer(s, combinedAPIServer)
				pfs.RegisterInternalApiServer(s, combinedAPIServer)
			}
		},
		func(t *testing.T, clientConns map[string]*grpc.ClientConn) {
			var clientConn *grpc.ClientConn
			for _, c := range clientConns {
				clientConn = c
				break
			}
			f(
				t,
				pfs.NewApiClient(
					clientConn,
				),
			)
		},
	)
}
