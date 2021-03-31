package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

var n, bytes int64

type adminServer struct {
	admin.UnimplementedAPIServer
}

func (_ adminServer) InspectCluster(ctx context.Context, req *types.Empty) (*admin.ClusterInfo, error) {
	return &admin.ClusterInfo{
		ID:           "foobar",
		DeploymentID: "foobar",
	}, nil
}

type pfsServer struct {
	pfs.UnimplementedAPIServer
}

func (_ pfsServer) PutFile(srv pfs.API_PutFileServer) error {
	atomic.AddInt64(&n, 1)
	defer atomic.AddInt64(&n, -1)
	for {
		msg, err := srv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err := srv.SendAndClose(&types.Empty{})
				return err
			}
			return err
		}
		atomic.AddInt64(&bytes, int64(len(msg.GetValue())))
	}
}

func main() {
	server, err := grpcutil.NewServer(context.Background(), false)
	if err != nil {
		log.Fatal(err)
	}
	admin.RegisterAPIServer(server.Server, &adminServer{})
	pfs.RegisterAPIServer(server.Server, &pfsServer{})
	server.ListenTCP("0.0.0.0", 30650)
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("%v recvd (%v open)\n", humanize.Bytes(uint64(atomic.LoadInt64(&bytes))), atomic.LoadInt64(&n))
		}
	}()
	server.Wait()
}
