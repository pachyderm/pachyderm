package collection

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc" // TODO: this is for db connection - move to a separate package
)

type FooModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primaryKey"`
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
	err := gc.WithLocalDB(func(db *gorm.DB) error {
		collection := NewPostgresCollection(db, &FooModel{})
		foos := collection.ReadOnly(context.Background())

		count, err := foos.Count()
		require.NoError(t, err)
		require.Equal(t, 0, count)
		return nil
	})
	require.NoError(t, err)
}
