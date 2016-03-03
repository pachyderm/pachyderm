package server

import (
	"fmt"
	"strings"
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
	pfsClient := pfs.NewAPIClient(clientConn)
	repo := uniqueString("TestSimple")
	err = pfsutil.CreateRepo(pfsClient, repo)
	require.NoError(t, err)
	_, err = pfsutil.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo"))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
