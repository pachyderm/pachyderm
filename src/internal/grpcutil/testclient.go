package grpcutil

import (
	"context"
	"net"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewTestClient(t testing.TB, regFunc func(*grpc.Server)) *grpc.ClientConn {
	ctx := pctx.TestContext(t)
	eg := errgroup.Group{}
	gserv := grpc.NewServer()
	listener := bufconn.Listen(1 << 20)
	regFunc(gserv)
	eg.Go(func() error {
		if err := gserv.Serve(listener); err != nil {
			log.Error(ctx, "gRPC server exited", zap.Error(err))
			return errors.EnsureStack(err)
		}
		return nil
	})
	gconn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { //nolint:SA1019
		res, err := listener.Dial()
		return res, errors.EnsureStack(err)
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() {
		gserv.Stop()
		err := eg.Wait()
		if errors.Is(err, grpc.ErrServerStopped) {
			err = nil
		}
		require.NoError(t, err)
	})
	return gconn
}
