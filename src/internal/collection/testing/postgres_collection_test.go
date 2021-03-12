package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type FooModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `collection:"primaryKey"`
	Value     string
}

func (fm *FooModel) TableName() string {
	return "foos"
}

func (fm *FooModel) WriteToProtobuf(val proto.Message) error {
	pb, ok := val.(*Foo)
	if !ok {
		return errors.Errorf("incorrect protobuf type")
	}
	pb.ID = fm.ID
	pb.Value = fm.Value
	return nil
}

func (fm *FooModel) LoadFromProtobuf(val proto.Message) error {
	pb, ok := val.(*Foo)
	if !ok {
		return errors.Errorf("incorrect protobuf type")
	}
	fm.ID = pb.ID
	fm.Value = pb.Value
	return nil
}

func TestPostgresCollections(suite *testing.T) {
	postgres := dbtesting.NewPostgresDeployment(suite)

	initCollection := func(t *testing.T) (*sqlx.DB, col.PostgresCollection) {
		db := postgres.NewDatabase(t)
		fooCol := col.NewPostgresCollection(db, &FooModel{})

		// Write several rows
		writeRow := func(i int) *Foo {
			foo := &Foo{ID: fmt.Sprintf("%d", i), Value: "old"}
			err := col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				readWrite := fooCol.ReadWrite(tx)
				return readWrite.Create(foo.ID, foo)
			})
			require.NoError(t, err)
			return foo
		}

		rows := []*Foo{}
		for i := 0; i < 10; i++ {
			rows = append(rows, writeRow(i))
		}

		return db, fooCol
	}

	suite.Run("ReadOnly", func(t *testing.T) {
		_, fooCol := initCollection(t)
		readOnly := fooCol.ReadOnly(context.Background())

		t.Run("Get", func(t *testing.T) {
			fooProto := Foo{}
			require.NoError(t, readOnly.Get("5", &fooProto))
			require.Equal(t, "5", fooProto.ID)
		})

		t.Run("NotFound", func(t *testing.T) {
			fooProto := Foo{}
			err := readOnly.Get("baz", &fooProto)
			require.True(t, errors.Is(err, col.ErrNotFound{}))
			require.True(t, col.IsErrNotFound(err))
		})

		t.Run("Count", func(t *testing.T) {
			count, err := readOnly.Count()
			require.NoError(t, err)
			require.Equal(t, int64(10), count)
		})
	})

	suite.Run("ReadWrite", func(t *testing.T) {
		suite.Run("Create", func(t *testing.T) {
			db, fooCol := initCollection(t)
			fooProto := Foo{}

			// Fails when overwriting a row
			err := col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				readWrite := fooCol.ReadWrite(tx)
				fooProto.ID = "5"
				return readWrite.Create(fooProto.ID, &fooProto)
			})
			require.YesError(t, err)
			require.True(t, col.IsErrExists(err))
			require.True(t, errors.Is(err, col.ErrExists{Type: "foos", Key: "5"}))

			// Succeeds when creating a new row
			err = col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				readWrite := fooCol.ReadWrite(tx)
				fooProto.ID = "10"
				return readWrite.Create(fooProto.ID, &fooProto)
			})
			require.NoError(t, err)

			// Fails when doing both
			err = col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				readWrite := fooCol.ReadWrite(tx)
				fooProto.ID = "11"
				if err := readWrite.Create(fooProto.ID, &fooProto); err != nil {
					return err
				}
				fooProto.ID = "6"
				return readWrite.Create(fooProto.ID, &fooProto)
			})
			require.YesError(t, err)
			require.True(t, col.IsErrExists(err))
			require.True(t, errors.Is(err, col.ErrExists{Type: "foos", Key: "6"}))

			// Check that only the correct rows exist
			readOnly := fooCol.ReadOnly(context.Background())

			expectedKeys := []string{}
			actualKeys := []string{}
			for i := 0; i <= 10; i++ {
				expectedKeys = append(expectedKeys, fmt.Sprintf("%d", i))
			}
			require.NoError(t, readOnly.List(&fooProto, col.DefaultOptions(), func() error {
				actualKeys = append(actualKeys, fooProto.ID)
				return nil
			}))
			require.ElementsEqual(t, expectedKeys, actualKeys)

			// Double-check that the count operation agrees
			count, err := readOnly.Count()
			require.NoError(t, err)
			require.Equal(t, int64(len(expectedKeys)), count)
		})

		suite.Run("Put", func(t *testing.T) {
			db, fooCol := initCollection(t)
			fooProto := Foo{}

			// Fails if row doesn't exist
			err := col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				readWrite := fooCol.ReadWrite(tx)
				fooProto.ID = "5"
				return readWrite.Create(fooProto.ID, &fooProto)
			})
			require.YesError(t, err)
			require.True(t, col.IsErrExists(err))
			require.True(t, errors.Is(err, col.ErrExists{Type: "foos", Key: "5"}))
		})

		suite.Run("Update", func(t *testing.T) {
		})

		suite.Run("Upsert", func(t *testing.T) {
		})

		suite.Run("Delete", func(t *testing.T) {
		})

		suite.Run("DeleteAll", func(t *testing.T) {
		})

		suite.Run("DeleteAllPrefix", func(t *testing.T) {
		})
	})

	// Watch tests
}
