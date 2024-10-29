package snapshotdb

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	snapshotpb "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateAndGetJob(t *testing.T) {
	withDependencies(t, func(d dependencies) {
		metadata := map[string]string{"key": "value"}
		id, err := CreateSnapshot(d.ctx, d.tx, d.s, metadata)
		if err != nil {
			t.Fatalf("create snapshot in database: %v", err)
		}
		snapshot, internal, err := GetSnapshot(d.ctx, d.tx, id)
		if err != nil {
			t.Fatalf("get snapshot %d from database: %v", id, err)
		}
		t.Log(snapshot, internal)
	})
}

func TestListSnapshotTxByFilter(t *testing.T) {
	ctx, db := DB(t)
	fs := FilesetStorage(t, db)
	var want []*snapshotpb.SnapshotInfo
	err := dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		_, err := CreateSnapshot(ctx, sqlTx, fs, map[string]string{})
		return errors.Wrap(err, "create snapshot")
	})
	if err != nil {
		t.Fatalf("with tx: %v", err)
	}
	// only snapshots 2 to 5 are expected
	since := time.Now()
	err = dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		want = nil
		for i := 0; i < 4; i++ {
			id, err := CreateSnapshot(ctx, sqlTx, fs, map[string]string{})
			if err != nil {
				return errors.Wrapf(err, "create snapshot in iteration %d: %v", i+1, err)
			}
			s, _, err := GetSnapshot(ctx, sqlTx, id)
			if err != nil {
				return errors.Wrapf(err, "get snapshot struct in iteration %d: %v", i+1, err)
			}
			want = append(want, s)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("with tx: %v", err)
	}
	err = dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		got, err := ListSnapshot(ctx, sqlTx, since, 0)
		if err != nil {
			t.Fatalf("list snapshot by filter: %v", err)
		}
		require.NoDiff(t, want, got, []cmp.Option{protocmp.Transform()})
		return nil
	})
	if err != nil {
		t.Fatalf("with tx: %v", err)
	}
}

func TestDeleteSnapshot(t *testing.T) {
	ctx, db := DB(t)
	fs := FilesetStorage(t, db)
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		_, err := CreateSnapshot(ctx, sqlTx, fs, map[string]string{})
		return errors.Wrap(err, "create snapshot")
	}); err != nil {
		t.Fatalf("with tx: %v", err)
	}
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		err := DeleteSnapshot(ctx, sqlTx, 1)
		return errors.Wrap(err, "delete snapshot")
	}); err != nil {
		t.Fatalf("with tx: %v", err)
	}
}
