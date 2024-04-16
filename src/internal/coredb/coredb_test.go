package coredb_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestClusterMetadata(t *testing.T) {
	var db *sqlx.DB
	pachd.NewTestPachd(t, pachd.TestPachdOption{
		MutateEnv: func(env *pachd.Env) {
			db = env.DB
		},
	})
	ctx := pctx.TestContext(t)
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		got, err := coredb.GetClusterMetadata(ctx, tx)
		if err != nil {
			t.Fatalf("initial get: %v", err)
		}
		want := map[string]string{}
		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("initial metadata read (-want +got):\n%s", diff)
		}
		want = map[string]string{"key": "value"}
		if err := coredb.UpdateClusterMetadata(ctx, tx, want); err != nil {
			t.Fatalf("update metadata: %v", err)
		}
		got, err = coredb.GetClusterMetadata(ctx, tx)
		if err != nil {
			t.Fatalf("get after update: %v", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("metadata after update (-want +got):\n%s", diff)
		}
		return nil
	}); err != nil {
		t.Fatalf("tx failed: %v", err)
	}
}
