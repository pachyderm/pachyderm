package pfstest

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
	"github.com/phayes/freeport"
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
	counter int32 = 0
)

func TestBtrfs(t *testing.T) {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_BTRFS_ROOT")
	if rootDir == "" {
		t.Fatal("PFS_BTRFS_ROOT not set")
	}
	runAllTests(t, drive.NewBtrfsDriver(rootDir), testDefaultNumShards)
}

func runAllTests(t *testing.T, driver drive.Driver, numShards int) {
	runTest(t, driver, numShards, testInit)
	runTest(t, driver, numShards, testInitGetPut)
}

func testInit(t *testing.T, apiClient pfs.ApiClient) {
	repositoryName := testRepositoryName()
	initRepositoryResponse, err := apiClient.InitRepository(
		context.Background(),
		&pfs.InitRepositoryRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, initRepositoryResponse)
}

func testInitGetPut(t *testing.T, apiClient pfs.ApiClient) {
	repositoryName := testRepositoryName()
	initRepositoryResponse, err := apiClient.InitRepository(
		context.Background(),
		&pfs.InitRepositoryRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, initRepositoryResponse)

	putFileResponse, err := apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: "scratch",
				},
				Path: "path/to/one",
			},
			Value: []byte("hello world"),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, putFileResponse)

	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: "scratch",
				},
				Path: "path/to/one",
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, apiGetFileClient)

	buffer := bytes.NewBuffer(nil)
	for bytesValue, err := apiGetFileClient.Recv(); err != io.EOF; bytesValue, err = apiGetFileClient.Recv() {
		require.NoError(t, err)
		_, writeErr := buffer.Write(bytesValue.Value)
		require.NoError(t, writeErr)
	}
	require.Equal(t, "hello world", buffer.String())
}

func testRepositoryName() string {
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}

func runTest(
	t *testing.T,
	driver drive.Driver,
	numShards int,
	f func(t *testing.T, apiClient pfs.ApiClient),
) {
	runGrpcTest(
		t,
		func(s *grpc.Server, a string) {
			pfs.RegisterApiServer(
				s,
				server.NewAPIServer(
					shard.NewSharder(
						numShards,
					),
					route.NewRouter(
						address.NewSingleAddresser(
							a,
						),
						dial.NewDialer(),
						a,
					),
					driver,
				),
			)
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
	errC         chan error
}

func (g *grpcSuite) SetupSuite() {
	port := freeport.GetPort()
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	require.NoError(g.T(), err)
	g.server = grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	g.registerFunc(g.server, address)
	g.errC = make(chan error, 1)
	go func() {
		g.errC <- g.server.Serve(listener)
		close(g.errC)
	}()
	clientConn, err := grpc.Dial(address)
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
