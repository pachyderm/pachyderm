package pfstest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/discovery"
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
	testShardsPerServer = 16
	testNumServers      = 16
	testSize            = 10000
)

var (
	counter int32
)

func init() {
	// TODO(pedge): needed in tests? will not be needed for golang 1.5 for sure
	runtime.GOMAXPROCS(runtime.NumCPU())
	executil.SetDebug(true)
}

func TestBtrfs(t *testing.T) {
	t.Parallel()
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

	var wg sync.WaitGroup
	for i := 0; i < testSize; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			iErr := putFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
			iErr = putFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
		}()
	}
	wg.Wait()

	err = commit(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	getCommitInfoResponse, err = getCommitInfo(apiClient, repositoryName, newCommitID)
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
			readStringer, iErr := getFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i))
			require.NoError(t, iErr)
			require.Equal(t, fmt.Sprintf("hello%d", i), readStringer.String())
			readStringer, iErr = getFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i))
			require.NoError(t, iErr)
			require.Equal(t, fmt.Sprintf("hello%d", i), readStringer.String())
		}()
	}
	wg.Wait()

	listFilesResponse, err := listFiles(apiClient, repositoryName, newCommitID, "a/b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))
	listFilesResponse, err = listFiles(apiClient, repositoryName, newCommitID, "a/c", 0, 1)
	require.NoError(t, err)
	require.Equal(t, testSize, len(listFilesResponse.FileInfo))

	var fileInfos [7][]*pfs.FileInfo
	wg = sync.WaitGroup{}
	for i := 0; i < 7; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			listFilesResponse, err = listFiles(apiClient, repositoryName, newCommitID, "a/b", i, 7)
			require.NoError(t, err)
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
			discoveryClient := discovery.NewMockClient()
			i := 0
			addresses := make([]string, testNumServers)
			for address := range servers {
				shards := make([]string, testShardsPerServer)
				for j := 0; j < testShardsPerServer; j++ {
					shards[j] = fmt.Sprintf("%d", (i*testShardsPerServer)+j)
				}
				_ = discoveryClient.Set(address+"-master", strings.Join(shards, ","))
				addresses[i] = address
				i++
			}
			_ = discoveryClient.Set("all-addresses", strings.Join(addresses, ","))
			for address, s := range servers {
				combinedAPIServer := server.NewCombinedAPIServer(
					route.NewSharder(
						testShardsPerServer*testNumServers,
					),
					route.NewRouter(
						route.NewDiscoveryAddresser(
							discoveryClient,
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
			for _, c := range clientConns {
				if c != clientConn {
					_ = c.Close()
				}
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
