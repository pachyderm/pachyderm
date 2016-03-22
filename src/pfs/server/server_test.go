package server

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
)

const (
	shards  = 32
	servers = 4
)

var (
	port int32 = 30651
)

func TestBlock(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pfsutil.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pfsutil.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pfsutil.NewDiff("foo", "c2", 0),
		})
	require.NoError(t, err)
	listDiffClient, err := blockClient.ListDiff(
		context.Background(),
		&pfs.ListDiffRequest{Shard: 0},
	)
	require.NoError(t, err)
	var diffInfos []*pfs.DiffInfo
	for {
		diffInfo, err := listDiffClient.Recv()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
		diffInfos = append(diffInfos, diffInfo)
	}
	require.Equal(t, 3, len(diffInfos))
}

func TestSimple(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := uniqueString("TestSimple")
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, repo, commit1.ID))
	commitInfos, err := pfsutil.ListCommit(pfsClient, []string{repo})
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsutil.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestBranch(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := uniqueString("TestBranch")
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := pfsutil.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit1, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
	commit2, err := pfsutil.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsutil.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsutil.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

func TestDisallowReadsDuringCommit(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := uniqueString("TestDisallowReadsDuringCommit")
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)

	// Make sure we can't get the file before the commit is finished
	var buffer bytes.Buffer
	require.YesError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, pfsutil.FinishCommit(pfsClient, repo, commit1.ID))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsutil.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func getBlockClient(t *testing.T) pfs.BlockAPIClient {
	localPort := atomic.AddInt32(&port, 1)
	address := fmt.Sprintf("localhost:%d", localPort)
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	blockAPIServer, err := NewLocalBlockAPIServer(root)
	require.NoError(t, err)
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pachyderm.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfs.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfs.APIServer,
	internalAPIServer pfs.InternalAPIServer, blockAPIServer pfs.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterInternalAPIServer(s, internalAPIServer)
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pachyderm.Version},
			protoserver.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClientAndServer(t *testing.T) (pfs.APIClient, []*internalAPIServer) {
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	var ports []int32
	for i := 0; i < servers; i++ {
		ports = append(ports, atomic.AddInt32(&port, 1))
	}
	var addresses []string
	for _, port := range ports {
		addresses = append(addresses, fmt.Sprintf("localhost:%d", port))
	}
	sharder := shard.NewLocalSharder(addresses, shards)
	var internalAPIServers []*internalAPIServer
	for i, port := range ports {
		address := addresses[i]
		driver, err := drive.NewDriver(address)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		hasher := pfs.NewHasher(shards, 1)
		dialer := grpcutil.NewDialer(grpc.WithInsecure())
		apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
		internalAPIServer := newInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
		internalAPIServers = append(internalAPIServers, internalAPIServer)
		runServers(t, port, apiServer, internalAPIServer, blockAPIServer)
		for i := 0; i < shards; i++ {
			require.NoError(t, internalAPIServer.AddShard(uint64(i)))
		}
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pfs.NewAPIClient(clientConn), internalAPIServers
}

func restartServer(servers []*internalAPIServer, t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, server := range servers {
		server := server
		for i := 0; i < shards; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				require.NoError(t, server.DeleteShard(uint64(i)))
				require.NoError(t, server.AddShard(uint64(i)))
			}()
		}
	}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
