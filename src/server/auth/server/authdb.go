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
	oidcStatesCollectionName   = "oidc_states"
)

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(authConfigCollectionName, nil, nil, nil, nil, nil),
		col.NewPostgresCollection(roleBindingsCollectionName, nil, nil, nil, nil, nil),
		col.NewPostgresCollection(membersCollectionName, nil, nil, nil, nil, nil),
		col.NewPostgresCollection(groupsCollectionName, nil, nil, nil, nil, nil),
		col.NewPostgresCollection(oidcStatesCollectionName, nil, nil, nil, nil, nil),
	}
}

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

func roleBindingsCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		roleBindingsCollectionName,
		db,
		listener,
		&auth.RoleBinding{},
		nil,
		nil,
	)
}

func membersCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		membersCollectionName,
		db,
		listener,
		&auth.Groups{},
		nil,
		nil,
	)
}

func groupsCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		groupsCollectionName,
		db,
		listener,
		&auth.Users{},
		nil,
		nil,
	)
}

func oidcStatesCollection(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		oidcStatesCollectionName,
		db,
		listener,
		&auth.SessionInfo{},
		nil,
		nil,
	)
}
