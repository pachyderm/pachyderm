package server

import (
	"fmt"
	"runtime/pprof"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

func NewDebugServer(etcdClient *etcd.Client) debug.DebugServer {
	return &DebugServer{
		etcdClient: etcdClient,
	}
}

type DebugServer struct {
	etcdClient *etcd.Client
}

func (s *DebugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("unable to find goroutine profile")
	}
	if err := profile.WriteTo(grpcutil.NewStreamingBytesWriter(server), 2); err != nil {
		return err
	}
	if request.Recurse {
	}
	return nil
}
