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
	port   = 30650
	shards = 1
)

func TestSimple(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
	repo := uniqueString("TestSimple")
	err := pfsutil.CreateRepo(pfsClient, repo)
	require.NoError(t, err)
	commit, err := pfsutil.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = pfsutil.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer)
	require.NoError(t, err)
	require.Equal(t, "foo\n", buffer.String())
	branches, err := pfsutil.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

var client pfs.APIClient
var clientOnce sync.Once

func getPfsClient(t *testing.T) pfs.APIClient {
	clientOnce.Do(func() {
		address := fmt.Sprintf("localhost:%d", port)
		driver, err := drive.NewDriver(address)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(uniqueString("/tmp/pach_test/run"))
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
