package server

import (
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
)

const (
	authConfigCollectionName   = "auth_config"
	roleBindingsCollectionName = "role_bindings"
	membersCollectionName      = "members"
	groupsCollectionName       = "groups"
)

var authConfigIndexes = []*col.Index{}

func authConfigCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		authConfigCollectionName,
		db,
		listener,
		&auth.OIDCConfig{},
		nil,
		nil,
	)
}

var roleBindingsIndexes = []*col.Index{}

func roleBindingsCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		roleBindingsCollectionName,
		db,
		listener,
		&auth.RoleBinding{},
		roleBindingsIndexes,
		nil,
	)
}

var membersIndexes = []*col.Index{}

func membersCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		membersCollectionName,
		db,
		listener,
		&auth.Groups{},
		membersIndexes,
		nil,
	)
}

var groupsIndexes = []*col.Index{}

func groupsCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		groupsCollectionName,
		db,
		listener,
		&auth.Users{},
		groupsIndexes,
		nil,
	)
}

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(authConfigCollectionName, nil, nil, nil, authConfigIndexes, nil),
		col.NewPostgresCollection(roleBindingsCollectionName, nil, nil, nil, roleBindingsIndexes, nil),
		col.NewPostgresCollection(membersCollectionName, nil, nil, nil, membersIndexes, nil),
		col.NewPostgresCollection(groupsCollectionName, nil, nil, nil, groupsIndexes, nil),
	}
}
