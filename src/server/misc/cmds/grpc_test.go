package cmds

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

type fakeAdmin struct {
	admin.UnimplementedAPIServer
}

func (fakeAdmin) InspectCluster(ctx context.Context, req *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
	if req.GetClientVersion().GetMajor() != 42 {
		return nil, status.Error(codes.InvalidArgument, "wrong version")
	}
	info := new(admin.ClusterInfo)
	info.WarningsOk = true
	return info, nil
}

func TestGRPC(t *testing.T) {
	ctx := pctx.TestContext(t)
	s := grpc.NewServer()
	admin.RegisterAPIServer(s, new(fakeAdmin))
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		if err := s.Serve(l); err != nil {
			panic(err)
		}
	}()
	t.Cleanup(s.Stop)
	out := new(bytes.Buffer)
	p := gRPCParams{Address: l.Addr().String()}
	if err := p.Run(ctx, nil, out, []string{"admin_v2.API.InspectCluster", `{"clientVersion":{"major":42}}`}); err != nil {
		t.Fatalf("rpc failed: %v", err)
	}
	got, want := &admin.ClusterInfo{}, &admin.ClusterInfo{WarningsOk: true}
	if err := protojson.Unmarshal(out.Bytes(), got); err != nil {
		t.Fatalf("unmarshal reply: %v", err)
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("reply (-want +got):\n%v", err)
	}
}
