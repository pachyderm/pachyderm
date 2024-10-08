package snapshotdb_test

import (
	"encoding/json"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"testing"
	"time"
)

func TestCreateAndGetJob(t *testing.T) {
	withDependencies(t, func(d dependencies) {
		metadata := map[string]string{"key": "value"}
		jsonData, err := json.Marshal(metadata)
		if err != nil {
			t.Fatalf("marshal metadata: %v", err)
		}
		id, err := snapshotdb.CreateSnapshot(d.ctx, d.tx, d.s, jsonData)
		if err != nil {
			t.Fatalf("create snapshot in database: %v", err)
		}
		snpshot, err := snapshotdb.GetSnapshot(d.ctx, d.tx, id)
		if err != nil {
			t.Fatalf("get snapshot %d from database: %v", id, err)
		}
		t.Log(snpshot)
	})
}

func TestListSnapshotTxByFilter(t *testing.T) {
	ctx, db := DB(t)
	fs := FilesetStorage(t, db)
	expected := make([]snapshotdb.Snapshot, 0)
	withTx(t, ctx, db, fs, func(d dependencies) {
		_, err := snapshotdb.CreateSnapshot(d.ctx, d.tx, d.s, []byte{})
		if err != nil {
			t.Fatalf("create snapshot in iteration 1: %v", err)
		}
	})
	time.Sleep(3 * time.Second)
	withTx(t, ctx, db, fs, func(d dependencies) {
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
	})
	// only snapshots 2 to 5 are expected
	createdAfter := time.Now().Add(-1 * time.Second)
	withTx(t, ctx, db, fs, func(d dependencies) {
		snapshots, err := snapshotdb.ListSnapshotTxByFilter(d.ctx, d.tx,
			snapshotdb.IterateSnapshotsRequest{
				Filter: snapshotdb.IterateSnapshotsFilter{
					CreatedAfter: createdAfter,
				},
			})
		if err != nil {
			t.Fatalf("list snapshot by filter: %v", err)
		}
		require.NoDiff(t, expected, snapshots, nil)
	})
}
