package snapshot

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := NewTestClient(t)
	metadata := map[string]string{"key": "value"}
	resp, err := c.CreateSnapshot(ctx, &snapshot.CreateSnapshotRequest{
		Metadata: metadata,
	})
	if err != nil {
		t.Fatalf("create snapshot RPC: %v", err)
	}
	t.Log(resp)
}

func TestListSnapshot(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := NewTestClient(t)
	createSnapshots(t, ctx, c)
	time.Sleep(2 * time.Second)
	since := time.Now().Add(-1 * time.Second)
	createSnapshots(t, ctx, c)
	listClient, err := c.ListSnapshot(ctx, &snapshot.ListSnapshotRequest{
		Since: timestamppb.New(since),
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("list snapshot RPC: %v", err)
	}
	expect := []int64{6, 7, 8}
	var actual []int64
	for {
		resp, err := listClient.Recv()
		if err != nil {
			if errors.As(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
		actual = append(actual, resp.Info.Id)
	}
	require.NoDiff(t, expect, actual, nil)
}

func createSnapshots(t *testing.T, ctx context.Context, c snapshot.APIClient) {
	for i := 0; i < 5; i++ {
		_, err := c.CreateSnapshot(ctx, &snapshot.CreateSnapshotRequest{})
		if err != nil {
			t.Fatalf("create snapshot RPC in iteration %d: %v", i, err)
		}
	}
}
