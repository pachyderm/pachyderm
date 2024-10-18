package snapshot_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	snapshotpb "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	createSnapshots(t, ctx, c)
}

func TestListSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	createSnapshots(t, ctx, c)
	// get all snapshots
	listClient, err := c.ListSnapshot(ctx, &snapshot.ListSnapshotRequest{})
	if err != nil {
		t.Fatalf("list snapshot RPC: %v", err)
	}
	allRows, err := grpcutil.Collect[*snapshotpb.ListSnapshotResponse](listClient, 100)
	if err != nil {
		t.Fatalf("grpcutil collect list response: %v", err)
	}
	// get snapshots from the second-earliest one and limit 3
	since := allRows[3].Info.CreatedAt
	listClient2, err := c.ListSnapshot(ctx, &snapshot.ListSnapshotRequest{
		Limit: 3,
		Since: since,
	})
	if err != nil {
		t.Fatalf("list snapshot RPC: %v", err)
	}
	sinceSecondSnapshot, err := grpcutil.Collect[*snapshotpb.ListSnapshotResponse](listClient2, 100)
	if err != nil {
		t.Fatalf("grpcutil collect list response: %v", err)
	}
	// the returned snapshots are in desc order.
	// So sinceSecondSnapshot should be 5,4,3 and allRows are 5,4,3,2,1
	require.NoDiff(t, allRows[:3], sinceSecondSnapshot, []cmp.Option{protocmp.Transform()})
}

func TestInspectSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	createResp, err := c.CreateSnapshot(ctx, &snapshot.CreateSnapshotRequest{})
	if err != nil {
		t.Fatalf("create snapshot RPC: %v", err)
	}
	inspectResp, err := c.InspectSnapshot(ctx, &snapshot.InspectSnapshotRequest{Id: createResp.Id})
	if err != nil {
		t.Fatalf("inspect snapshot RPC: %v", err)
	}
	got := inspectResp.Info
	want := &snapshot.SnapshotInfo{
		Id:               createResp.Id,
		ChunksetId:       1,
		CreatedAt:        inspectResp.Info.CreatedAt, // the created time is not compared
		PachydermVersion: version.Version.String(),
	}
	require.NoDiff(t, want, got, []cmp.Option{protocmp.Transform()})
}

func TestDeleteSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	createResp, err := c.CreateSnapshot(ctx, &snapshot.CreateSnapshotRequest{})
	if err != nil {
		t.Fatalf("create snapshot RPC: %v", err)
	}
	_, err = c.DeleteSnapshot(ctx, &snapshot.DeleteSnapshotRequest{Id: createResp.Id})
	if err != nil {
		t.Fatalf("delete snapshot RPC: %v", err)
	}
	// sanity check. we cannot inspect the deleted snapshot
	_, err = c.InspectSnapshot(ctx, &snapshot.InspectSnapshotRequest{Id: createResp.Id})
	if err == nil || status.Convert(err).Code() != codes.NotFound {
		t.Fatalf("want error code not found")
	}
}

func createSnapshots(t *testing.T, ctx context.Context, c snapshot.APIClient) {
	for i := 0; i < 5; i++ {
		resp, err := c.CreateSnapshot(ctx, &snapshot.CreateSnapshotRequest{})
		if err != nil {
			t.Fatalf("create snapshot RPC in iteration %d: %v", i, err)
		}
		if resp.Id == 0 {
			t.Fatalf("id should be 1, got %d", resp.Id)
		}
	}
}
