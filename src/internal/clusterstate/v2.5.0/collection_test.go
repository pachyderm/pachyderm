package v2_5_0

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var (
	TestSecondaryIndex = &col.Index{
		Name: "Value",
		Extract: func(val proto.Message) string {
			return val.(*col.TestItem).Value
		},
	}
)

func newTestDB(t testing.TB) (*pachsql.DB, string) {
	options := dockertestenv.NewTestDBOptions(t)
	dsn := dbutil.GetDSN(options...)
	db, err := dbutil.NewDB(options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db, dsn
}

// TestMigratePostgreSQLCollection creates a simple collection with two items,
// then migrates them both to have a new key and mutate one member, and finally
// verifies the existence of the new items and the non-existence of the old.
func TestMigratePostgreSQLCollection(t *testing.T) {
	db, dsn := newTestDB(t)
	listener := col.NewPostgresListener(dsn)
	testCol := col.NewPostgresCollection("test_items", db, listener, &col.TestItem{}, []*col.Index{TestSecondaryIndex})
	ctx := pctx.TestContext(t)
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := col.CreatePostgresSchema(ctx, tx); err != nil {
			return err
		}
		if err := col.SetupPostgresV0(ctx, tx); err != nil {
			return err
		}
		return col.SetupPostgresCollections(ctx, tx, testCol)
	}); err != nil {
		t.Fatal("could create test collection:", err)
	}
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := testCol.ReadWrite(tx).Put("foo1", &col.TestItem{Id: "foo", Value: "bar", Data: "baz"}); err != nil {
			return errors.Wrap(err, "could not create first item")
		}
		return testCol.ReadWrite(tx).Put("bar2", &col.TestItem{Id: "bar", Value: "bim", Data: "test"})
	}); err != nil {
		t.Fatal("could not write test item:", err)
	}
	var indices = []*index{{TestSecondaryIndex.Name, TestSecondaryIndex.Extract}}
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		var oldItem = new(col.TestItem)
		return migratePostgreSQLCollection(ctx, tx, "test_items", indices, oldItem, func(oldKey string) (newKey string, newVal proto.Message, err error) {
			oldItem.Value = oldItem.Value + " quux"
			return oldKey + "new", oldItem, nil
		})
	}); err != nil {
		t.Fatal("could not migrate test item:", err)
	}
	var item col.TestItem
	if err := testCol.ReadOnly(ctx).Get("foo1new", &item); err != nil {
		t.Error("could not read migrated item:", err)
	}
	if item.Id != "foo" {
		t.Errorf("%q ≠ %q", item.Id, "foo")
	}
	if item.Value != "bar quux" {
		t.Errorf("%q ≠ %q", item.Value, "bar quux")
	}
	if err := testCol.ReadOnly(ctx).Get("foo1", &item); err != nil {
		if !col.IsErrNotFound(err) {
			t.Error("could not try to get migrated item:", err)
		}
	} else {
		t.Errorf("found migrated item under old key: %v", item)
	}
}
