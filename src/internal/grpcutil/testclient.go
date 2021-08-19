package grpcutil

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func NewTestClient(t testing.TB, regFunc func(*grpc.Server)) *grpc.ClientConn {
	ctx := context.Background()
	eg := errgroup.Group{}
	gserv := grpc.NewServer()
	listener := bufconn.Listen(1 << 20)
	regFunc(gserv)
	eg.Go(func() error {
		return gserv.Serve(listener)
	})
	gconn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() {
		gserv.GracefulStop()
		require.Nil(t, eg.Wait())
	})
	return gconn
}
