package server

import (
	"fmt"
	"testing"

	"go.pedge.io/pkg/http"
	"go.pedge.io/proto/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
)

const (
	port   = 30650
	shards = 32
)

func TestSimple(t *testing.T) {
	address := fmt.Sprintf("localhost:%d", port)
	driver, err := drive.NewDriver(address)
	require.NoError(t, err)
	blockAPIServer, err := NewLocalBlockAPIServer(uniqueString("/tmp/pach_test/"))
	require.NoError(t, err)
	sharder := shard.NewLocalSharder(address, shards)
	hasher := pfs.NewHasher(shards, 1)
	dialer := grpcutil.NewDialer(grpc.WithInsecure())
	apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
	internalAPIServer := NewInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
	go func() {
		err := protoserver.ServeWithHTTP(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterInternalAPIServer(s, internalAPIServer)
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
			},
			func(ctx context.Context, mux *runtime.ServeMux, clientConn *grpc.ClientConn) error {
				return pfs.RegisterAPIHandler(ctx, mux, clientConn)
			},
			protoserver.ServeWithHTTPOptions{
				ServeOptions: protoserver.ServeOptions{
					Version: pachyderm.Version,
				},
			},
			protoserver.ServeEnv{
				GRPCPort: port,
			},
			pkghttp.HandlerEnv{
			//Port: appEnv.HTTPPort,
			},
		)
		require.NoError(t, err)
	}()
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
