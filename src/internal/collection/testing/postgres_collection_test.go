package testing

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type FooModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `collection:"primaryKey"`
	Name      string
	Value     string
}

func (fm *FooModel) TableName() string {
	return "foos"
}

func (fm *FooModel) WriteToProtobuf(val proto.Message) error {
	return nil
}

func TestPostgresBasic(t *testing.T) {
	postgres := dbtesting.NewPostgresDeployment(t)
	db := postgres.NewDatabase(t)

	fooCol := col.NewPostgresCollection(db, &FooModel{})
	readOnly := fooCol.ReadOnly(context.Background())

	fooProto := Foo{}
	require.NoError(t, readOnly.Get("bar", &fooProto))

}
