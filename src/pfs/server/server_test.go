package server

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/discovery"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/pachyderm/pachyderm/src/pkg/grpctest"
	"github.com/stretchr/testify/require"
)

const (
	// TODO(pedge): large numbers of shards takes forever because
	// we are doing tons of btrfs operations on init, is there anything
	// we can do about that?
	testShardsPerServer = 8
	testNumServers      = 8
	testSize            = 1000
)

var (
	counter int32
)

func init() {
	// TODO(pedge): needed in tests? will not be needed for golang 1.5 for sure
	runtime.GOMAXPROCS(runtime.NumCPU())
	executil.SetDebug(true)
}

func TestBtrfsFFI(t *testing.T) {
	t.Parallel()
	driver := drive.NewBtrfsDriver(getBtrfsRootDir(t), btrfs.NewFFIAPI())
	runTest(t, driver, testSimple)
}

func TestBtrfsExec(t *testing.T) {
	t.Parallel()
	driver := drive.NewBtrfsDriver(getBtrfsRootDir(t), btrfs.NewExecAPI())
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

func testGetVersion(t *testing.T, apiClient pfs.ApiClient) {
	getVersionResponse, err := pfsutil.GetVersion(apiClient)
	require.NoError(t, err)
	require.Equal(t, common.VersionString(), pfs.VersionString(getVersionResponse.Version))
}

func testSimple(t *testing.T, apiClient pfs.ApiClient) {
	repositoryName := testRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName)
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
			_, iErr := pfsutil.PutFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
			require.NoError(t, iErr)
			_, iErr = pfsutil.PutFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i), strings.NewReader(fmt.Sprintf("hello%d", i)))
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
			reader, iErr := pfsutil.GetFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/b/file%d", i))
			require.NoError(t, iErr)
			buffer := bytes.NewBuffer(nil)
			_, iErr = buffer.ReadFrom(reader)
			require.NoError(t, iErr)
			require.Equal(t, fmt.Sprintf("hello%d", i), buffer.String())
			reader, iErr = pfsutil.GetFile(apiClient, repositoryName, newCommitID, fmt.Sprintf("a/c/file%d", i))
			require.NoError(t, iErr)
			buffer = bytes.NewBuffer(nil)
			_, iErr = buffer.ReadFrom(reader)
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
			listFilesResponse, iErr := pfsutil.ListFiles(apiClient, repositoryName, newCommitID, "a/b", i, 7)
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

func testRepositoryName() string {
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
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
			for address, server := range servers {
				combinedAPIServer := NewCombinedAPIServer(
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
				pfs.RegisterApiServer(server, combinedAPIServer)
				pfs.RegisterInternalApiServer(server, combinedAPIServer)
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
