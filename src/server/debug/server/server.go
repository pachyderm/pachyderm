package server

import (
	"fmt"
	"runtime/pprof"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/worker"
)

// NewDebugServer creates a new server that serves the debug api over GRPC
func NewDebugServer(name string, etcdClient *etcd.Client, etcdPrefix string) debug.DebugServer {
	return &debugServer{
		name:       name,
		etcdClient: etcdClient,
		etcdPrefix: etcdPrefix,
	}
}

type debugServer struct {
	name       string
	etcdClient *etcd.Client
	etcdPrefix string
}

func (s *debugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("unable to find goroutine profile")
	}
	w := grpcutil.NewStreamingBytesWriter(server)
	if s.name != "" {
		if _, err := fmt.Fprintf(w, "== %s ==\n\n", s.name); err != nil {
			return err
		}
	}
	if err := profile.WriteTo(w, 2); err != nil {
		return err
	}
	if !request.Recursed {
		request.Recursed = true
		cs, err := worker.Clients(server.Context(), "", s.etcdClient, s.etcdPrefix)
		if err != nil {
			return err
		}
		for _, c := range cs {
			if _, err := fmt.Fprintf(w, "\n"); err != nil {
				return err
			}
			dumpC, err := c.Dump(
				server.Context(),
				request,
			)
			if err != nil {
				return err
			}
			if err := grpcutil.WriteFromStreamingBytesClient(dumpC, w); err != nil {
				return err
			}
		}
	}
	return nil
}
