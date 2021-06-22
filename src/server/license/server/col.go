package server

import (
	"github.com/jmoiron/sqlx"

	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
)

const (
	licenseCollectionName = "license"
)

var licenseIndexes = []*col.Index{}

func licenseCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		licenseCollectionName,
		db,
		listener,
		&ec.LicenseRecord{},
		licenseIndexes,
		nil,
	)
}

func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		licenseCollection(nil, nil),
	}
}
