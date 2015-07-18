package pfstest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/facebookgo/freeport"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	// TODO(pedge): large numbers of shards takes forever because
	// we are doing tons of btrfs operations on init, is there anything
	// we can do about that?
	testDefaultNumShards = 4
)

var (
	counter int32
)

func init() {
	executil.SetDebug(true)
}

func TestBtrfs(t *testing.T) {
	driver := btrfs.NewDriver(getBtrfsRootDir(t))
	numShards := testDefaultNumShards
	runTest(t, driver, numShards, testSimple)
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

	err = putFile(apiClient, repositoryName, newCommitID, "a/b/one", strings.NewReader("hello world"))
	require.NoError(t, err)

	err = commit(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)

	getCommitInfoResponse, err = getCommitInfo(apiClient, repositoryName, newCommitID)
	require.NoError(t, err)
	require.NotNil(t, getCommitInfoResponse)
	require.Equal(t, newCommitID, getCommitInfoResponse.CommitInfo.Commit.Id)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, getCommitInfoResponse.CommitInfo.CommitType)
	require.Equal(t, "scratch", getCommitInfoResponse.CommitInfo.ParentCommit.Id)

	readStringer, err := getFile(apiClient, repositoryName, newCommitID, "a/b/one")
	require.NoError(t, err)
	require.Equal(t, "hello world", readStringer.String())

	listFilesResponse, err := listFiles(apiClient, repositoryName, newCommitID, "a/b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(listFilesResponse.FileInfo))
	fileInfo := listFilesResponse.FileInfo[0]
	require.Equal(t, "a/b/one", fileInfo.Path.Path)
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
	numShards int,
	f func(t *testing.T, apiClient pfs.ApiClient),
) {
	runGrpcTest(
		t,
		func(s *grpc.Server, address string) {
			combinedAPIServer :=
				server.NewCombinedAPIServer(
					route.NewSharder(
						numShards,
					),
					route.NewRouter(
						route.NewSingleAddresser(
							address,
							numShards,
						),
						route.NewDialer(),
						address,
					),
					driver,
				)
			pfs.RegisterApiServer(s, combinedAPIServer)
			pfs.RegisterInternalApiServer(s, combinedAPIServer)
		},
		func(t *testing.T, clientConn *grpc.ClientConn) {
			f(
				t,
				pfs.NewApiClient(
					clientConn,
				),
			)
		},
	)
}

func runGrpcTest(
	t *testing.T,
	registerFunc func(*grpc.Server, string),
	testFunc func(*testing.T, *grpc.ClientConn),
) {
	grpcSuite := &grpcSuite{
		registerFunc: registerFunc,
		testFunc:     testFunc,
	}
	suite.Run(t, grpcSuite)
}

type grpcSuite struct {
	suite.Suite
	registerFunc func(*grpc.Server, string)
	testFunc     func(*testing.T, *grpc.ClientConn)
	clientConn   *grpc.ClientConn
	server       *grpc.Server
	address      string
	errC         chan error
}

func (g *grpcSuite) SetupSuite() {
	port, err := freeport.Get()
	require.NoError(g.T(), err)
	g.address = fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	require.NoError(g.T(), err)
	g.server = grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	g.registerFunc(g.server, g.address)
	g.errC = make(chan error, 1)
	go func() {
		g.errC <- g.server.Serve(listener)
		close(g.errC)
	}()
	clientConn, err := grpc.Dial(g.address)
	if err != nil {
		g.server.Stop()
		<-g.errC
		require.NoError(g.T(), err)
	}
	g.clientConn = clientConn
}

func (g *grpcSuite) TearDownSuite() {
	g.server.Stop()
	<-g.errC
	_ = g.clientConn.Close()
}

func (g *grpcSuite) TestSuite() {
	g.testFunc(g.T(), g.clientConn)
}
