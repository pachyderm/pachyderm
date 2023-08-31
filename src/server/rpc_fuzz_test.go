//go:build unit_test

package server

import (
	"context"
	"net"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestEmptyRequests(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	protos := []protoreflect.FileDescriptor{
		admin.File_admin_admin_proto,
		auth.File_auth_auth_proto,
		debug.File_debug_debug_proto,
		enterprise.File_enterprise_enterprise_proto,
		identity.File_identity_identity_proto,
		license.File_license_license_proto,
		pfs.File_pfs_pfs_proto,
		pps.File_pps_pps_proto,
		proxy.File_proxy_proxy_proto,
		transaction.File_transaction_transaction_proto,
		versionpb.File_version_versionpb_version_proto,
		worker.File_worker_worker_proto,
	}

	// TODO(jrockway): use pachClient.ClientConn() when that is merged
	cc, err := grpc.DialContext(ctx, net.JoinHostPort(env.PachClient.GetAddress().Host, strconv.Itoa(int(env.PachClient.GetAddress().Port))), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial real env: %v", err)
	}

	for _, fd := range protos {
		svcs := fd.Services()
		for si := 0; si < svcs.Len(); si++ {
			sd := svcs.Get(si)
			methods := sd.Methods()
			for mi := 0; mi < methods.Len(); mi++ {
				md := methods.Get(mi)
				name := string(sd.FullName()) + "." + string(md.Name())
				t.Run(name, func(t *testing.T) {
					ctx, c := context.WithTimeout(pctx.Child(ctx, name), 5*time.Second)
					testRPC(ctx, t, sd, md, cc, &emptypb.Empty{})
					defer c()
				})
			}
		}
	}
}

func testRPC(ctx context.Context, t *testing.T, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor, cc *grpc.ClientConn, req proto.Message) {
	fullName := "/" + path.Join(string(sd.FullName()), string(md.Name()))
	reply := &emptypb.Empty{}
	if err := cc.Invoke(ctx, fullName, req, reply); err != nil {
		t.Log(err)
	}
}
