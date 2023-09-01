//go:build unit_test

package server

import (
	"context"
	"net"
	"path"
	"strconv"
	"strings"
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
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var protos = []protoreflect.FileDescriptor{
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

func rangeRPCs(f func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor)) {
	for _, fd := range protos {
		svcs := fd.Services()
		for si := 0; si < svcs.Len(); si++ {
			sd := svcs.Get(si)
			methods := sd.Methods()
			for mi := 0; mi < methods.Len(); mi++ {
				md := methods.Get(mi)
				f(fd, sd, md)
			}
		}
	}
}

func TestEmptyRequests(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateAuthClient(t, env.PachClient, peerPort)

	// TODO(jrockway): use pachClient.ClientConn() when that is merged
	cc, err := grpc.DialContext(ctx, net.JoinHostPort(env.PachClient.GetAddress().Host, strconv.Itoa(int(env.PachClient.GetAddress().Port))), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial real env: %v", err)
	}

	rangeRPCs(func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) {
		name := string(sd.FullName()) + "." + string(md.Name())
		t.Run(name, func(t *testing.T) {
			switch {
			case strings.Contains(name, "RunLoadTest"):
				t.Skip("skipping load tests")
			case strings.Contains(name, "Deactivate"):
				t.Skip("skipping auth deactivation")
			case strings.Contains(name, "RotateRootToken"):
				t.Skip("skipping RotateRootToken")
			}
			ctx, c := context.WithTimeout(pctx.Child(ctx, name), 5*time.Second)
			defer c()
			client := env.PachClient.WithCtx(ctx)
			testRPC(client.Ctx(), t, sd, md, cc, &emptypb.Empty{})
		})

	})
}

func testRPC(ctx context.Context, t *testing.T, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor, cc *grpc.ClientConn, req proto.Message) {
	fullName := "/" + path.Join(string(sd.FullName()), string(md.Name()))
	reply := &emptypb.Empty{}
	if err := cc.Invoke(ctx, fullName, req, reply); err != nil {
		t.Log(err)
		if s, ok := status.FromError(err); ok {
			switch {
			case strings.Contains(s.Message(), "pachd mock"):
				t.Skip("skipping method that has no mock")
			case strings.Contains(s.Message(), "not activated"):
				t.Fatal("auth not activated?")
			case s.Code() == codes.Unimplemented:
				t.Skip("skipping unimplemented method/service")
			case s.Code() == codes.Unauthenticated:
				t.Fatal("unauthenticated?")
			}
		}
	}
}
