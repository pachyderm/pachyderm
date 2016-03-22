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

	pclient "github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/shard"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	shards = 1
)

var (
	port int32 = 30651
)

func TestBlock(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfsserver.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsserver.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsserver.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c2", 0),
		})
	require.NoError(t, err)
	listDiffClient, err := blockClient.ListDiff(
		context.Background(),
		&pfsclient.ListDiffRequest{Shard: 0},
	)
	require.NoError(t, err)
	var diffInfos []*pfsserver.DiffInfo
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
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	commitInfos, err := pfsclient.ListCommit(pfsClient, []string{repo})
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestBranch(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := uniqueString("TestBranch")
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit1, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
	commit2, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

func TestDisallowReadsDuringCommit(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := uniqueString("TestDisallowReadsDuringCommit")
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)

	// Make sure we can't get the file before the commit is finished
	var buffer bytes.Buffer
	require.YesError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func getBlockClient(t *testing.T) pfsclient.BlockAPIClient {
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
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfsclient.NewBlockAPIClient(clientConn)
}

func getClientAndServer(t *testing.T) (pfsclient.APIClient, *internalAPIServer) {
	localPort := atomic.AddInt32(&port, 1)
	address := fmt.Sprintf("localhost:%d", localPort)
	driver, err := drive.NewDriver(address)
	require.NoError(t, err)
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	blockAPIServer, err := NewLocalBlockAPIServer(root)
	require.NoError(t, err)
	sharder := shard.NewLocalSharder(address, shards)
	hasher := pfsserver.NewHasher(shards, 1)
	dialer := grpcutil.NewDialer(grpc.WithInsecure())
	apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
	internalAPIServer := newInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterAPIServer(s, apiServer)
				pfsclient.RegisterInternalAPIServer(s, internalAPIServer)
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	for i := 0; i < shards; i++ {
		require.NoError(t, internalAPIServer.AddShard(uint64(i)))
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	return pfsclient.NewAPIClient(clientConn), internalAPIServer
}

func restartServer(server *internalAPIServer, t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
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

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
