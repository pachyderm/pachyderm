package testing

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type TestModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `collection:"primaryKey"`
	Value     string
}

func (fm *TestModel) TableName() string {
	return "foos"
}

func (fm *TestModel) WriteToProtobuf(val proto.Message) error {
	pb, ok := val.(*TestItem)
	if !ok {
		return errors.Errorf("incorrect protobuf type")
	}
	pb.ID = fm.ID
	pb.Value = fm.Value
	return nil
}

func (fm *TestModel) LoadFromProtobuf(val proto.Message) error {
	pb, ok := val.(*TestItem)
	if !ok {
		return errors.Errorf("incorrect protobuf type")
	}
	fm.ID = pb.ID
	fm.Value = pb.Value
	return nil
}

func TestPostgresCollections(suite *testing.T) {
	suite.Parallel()
	postgres := dbtesting.NewPostgresDeployment(suite)

	newCollection := func(t *testing.T) (col.ReadOnlyCollection, WriteCallback) {
		db := postgres.NewDatabase(t)
		testCol := col.NewPostgresCollection(db, &TestModel{})

		writeCallback := func(f func(col.ReadWriteCollection) error) error {
			return col.NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
				return f(testCol.ReadWrite(tx))
			})
		}

		return testCol.ReadOnly(context.Background()), writeCallback
	}

	collectionTests(suite, newCollection)

	// TODO: postgres-specific collection tests
}
