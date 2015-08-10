package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpctest"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
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
	common.ForceLogColors()
}

func TestBtrfsFFI(t *testing.T) {
	t.Skip()
	t.Parallel()
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	runTest(t, driver, testSimple)
}

func TestBtrfsExec(t *testing.T) {
	t.Skip()
	t.Parallel()
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	runTest(t, driver, testSimple)
}

func TestFuseMount(t *testing.T) {
	t.Parallel()
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	runTest(t, driver, testMount)
}

func TestFuseMountBig(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	runTest(t, driver, testMountBig)

}

func BenchmarkFuse(b *testing.B) {
	driver := btrfs.NewDriver(getBtrfsRootDir(b))
	runBench(b, driver, benchMount)
}

func getBtrfsRootDir(tb testing.TB) string {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_BTRFS_ROOT")
	if rootDir == "" {
		tb.Fatal("PFS_BTRFS_ROOT not set")
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

func testMount(t *testing.T, apiClient pfs.ApiClient) {
	repositoryName := testRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName)
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

func testMountBig(t *testing.T, apiClient pfs.ApiClient) {
	if testing.Short() {
		return
	}
	repositoryName := testRepositoryName()

	err := pfsutil.InitRepository(apiClient, repositoryName)
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
	repositoryName := testRepositoryName()

	if err := pfsutil.InitRepository(apiClient, repositoryName); err != nil {
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

func testRepositoryName() string {
	// TODO could be nice to add callee to this string to make it easy to
	// recover results for debugging
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}

func testNamespace() string {
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}

func registerFunc(driver drive.Driver, discoveryClient discovery.Client, servers map[string]*grpc.Server) {
	addresser := route.NewDiscoveryAddresser(
		discoveryClient,
		testNamespace(),
	)
	i := 0
	for address := range servers {
		for j := 0; j < testShardsPerServer; j++ {
			// TODO(pedge): error
			_ = addresser.SetMasterAddress((i*testShardsPerServer)+j, address, 0)
		}
		i++
	}
	for address, server := range servers {
		combinedAPIServer := NewCombinedAPIServer(
			route.NewSharder(
				testShardsPerServer*testNumServers,
			),
			route.NewRouter(
				addresser,
				grpcutil.NewDialer(),
				address,
			),
			driver,
		)
		pfs.RegisterApiServer(server, combinedAPIServer)
		pfs.RegisterInternalApiServer(server, combinedAPIServer)
	}
}

func runTest(
	t *testing.T,
	driver drive.Driver,
	f func(t *testing.T, apiClient pfs.ApiClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(t, err)
	grpctest.Run(
		t,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			registerFunc(driver, discoveryClient, servers)
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

func runBench(
	b *testing.B,
	driver drive.Driver,
	f func(b *testing.B, apiClient pfs.ApiClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(b, err)
	grpctest.RunB(
		b,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			registerFunc(driver, discoveryClient, servers)
		},
		func(b *testing.B, clientConns map[string]*grpc.ClientConn) {
			var clientConn *grpc.ClientConn
			for _, c := range clientConns {
				clientConn = c
				break
			}
			f(
				b,
				pfs.NewApiClient(
					clientConn,
				),
			)
		},
	)
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
