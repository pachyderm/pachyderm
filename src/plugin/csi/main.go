package csi

import (
	"flag"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-hostpath", "name of the driver")
	nodeID     = flag.String("nodeid", "", "node id")
	port       = flag.Int("port", 9000, "The port on which this CSI plugin's services are served")
)

func main() {
	// TODO(msteffen) use src/client/pkg/grpcutil for this. Currently, I can't
	// because of vendored type mismatches (src/client/vendor/grpc.Server vs
	// src/plugin/csi/vendor/grpc.Server)
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint32),
		grpc.MaxRecvMsgSize(grpcutil.MaxMsgSize),
		grpc.MaxSendMsgSize(grpcutil.MaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	s := grpc.NewServer(opts...)
	csi.RegisterIdentityServer(s, &identitySvc{})
	csi.RegisterNodeServer(s, NewNodeSvc())
	csi.RegisterControllerServer(s, &controllerSvc{})
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(fmt.Sprintf("could not listen on %d: %v", *port, err))
	}
	if s.Serve(listener); err != nil {
		panic(fmt.Sprintf("error running CSI services: %v", err))
	}
}
