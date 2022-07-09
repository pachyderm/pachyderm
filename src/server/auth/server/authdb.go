package server

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

const (
	authConfigCollectionName   = "auth_config"
	roleBindingsCollectionName = "role_bindings"
	membersCollectionName      = "members"
	groupsCollectionName       = "groups"
)

var authConfigIndexes = []*col.Index{}

func authConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		authConfigCollectionName,
		db,
		listener,
		&auth.OIDCConfig{},
		nil,
	)
}

var roleBindingsIndexes = []*col.Index{}

func roleBindingsCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		roleBindingsCollectionName,
		db,
		listener,
		&auth.RoleBinding{},
		roleBindingsIndexes,
	)
}

var membersIndexes = []*col.Index{}

func membersCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		membersCollectionName,
		db,
		listener,
		&auth.Groups{},
		membersIndexes,
	)
}

var groupsIndexes = []*col.Index{}

func groupsCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		groupsCollectionName,
		db,
		listener,
		&auth.Users{},
		groupsIndexes,
	)
}

// CollectionsV0 returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(authConfigCollectionName, nil, nil, nil, authConfigIndexes),
		col.NewPostgresCollection(roleBindingsCollectionName, nil, nil, nil, roleBindingsIndexes),
		col.NewPostgresCollection(membersCollectionName, nil, nil, nil, membersIndexes),
		col.NewPostgresCollection(groupsCollectionName, nil, nil, nil, groupsIndexes),
	}
}

// InternalAuthUserPermissions adds the Internal Auth User as a cluster admin
func InternalAuthUserPermissions(tx *pachsql.Tx) error {
	roleBindings := roleBindingsCollection(nil, nil)
	var binding auth.RoleBinding
	if err := roleBindings.ReadWrite(tx).Get(clusterRoleBindingKey, &binding); err != nil {
		if col.IsErrNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "getting the cluster role binding")
	}
	binding.Entries[internalUser] = &auth.Roles{Roles: map[string]bool{auth.ClusterAdminRole: true}}
	return errors.EnsureStack(roleBindings.ReadWrite(tx).Put(clusterRoleBindingKey, &binding))
}
