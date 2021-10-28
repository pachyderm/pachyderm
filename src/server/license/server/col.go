package server

import (
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

const (
	licenseCollectionName = "license"
)

var licenseIndexes = []*col.Index{}

func licenseCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		licenseCollectionName,
		db,
		listener,
		&ec.LicenseRecord{},
		licenseIndexes,
	)
}

// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV0() []col.PostgresCollection {
	return []col.PostgresCollection{
		licenseCollection(nil, nil),
	}
}
