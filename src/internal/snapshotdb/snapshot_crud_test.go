package snapshotdb_test

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"testing"
	"time"
)

func TestListSnapshotTxByFilter(t *testing.T) {
	withDependencies(t, func(d dependencies) {
		expected := make([]snapshot.Snapshot, 0)

		_, err := snapshotdb.CreateSnapshot(d.ctx, d.tx, d.s, []byte{})
		if err != nil {
			t.Fatalf("create snapshot in iteration 1: %v", err)
		}
		createdAfter := time.Now()
		for i := 0; i < 4; i++ {
			id, err := snapshotdb.CreateSnapshot(d.ctx, d.tx, d.s, []byte{})
			if err != nil {
				t.Fatalf("create snapshot in iteration %d: %v", i+1, err)
			}
			s, err := snapshotdb.GetSnapshot(d.ctx, d.tx, id)
			if err != nil {
				t.Fatalf("get snapshot struct in iteration %d: %v", i+1, err)
			}
			expected = append(expected, s)
		}
		snapshots, err := snapshotdb.ListSnapshotTxByFilter(d.ctx, d.tx,
			snapshotdb.IterateSnapshotsRequest{
				Filter: snapshotdb.IterateSnapshotsFilter{
					CreatedAfter: createdAfter,
				},
			})
		require.NoError(t, err)
		require.NoDiff(t, expected, snapshots, nil)
	})
}
