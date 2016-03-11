package server

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

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
	port   = 30651
	shards = 32
)

func TestSimple(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
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
}

func TestBranch(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
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
}

func TestDisallowReadsDuringCommit(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
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
}

var client pfs.APIClient
var clientOnce sync.Once

func getPfsClient(t *testing.T) pfs.APIClient {
	clientOnce.Do(func() {
		address := fmt.Sprintf("localhost:%d", port)
		driver, err := drive.NewDriver(address)
		require.NoError(t, err)
		root := uniqueString("/tmp/pach_test/run")
		t.Logf("root %s", root)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		sharder := shard.NewLocalSharder(address, shards)
		hasher := pfs.NewHasher(shards, 1)
		dialer := grpcutil.NewDialer(grpc.WithInsecure())
		apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
		internalAPIServer := NewInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
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
				protoserver.ServeEnv{GRPCPort: port},
			)
			require.NoError(t, err)
		}()
		<-ready
		clientConn, err := grpc.Dial(address, grpc.WithInsecure())
		require.NoError(t, err)
		client = pfs.NewAPIClient(clientConn)
	})
	return client
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
