package snapshot_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	snapshotpb "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
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
