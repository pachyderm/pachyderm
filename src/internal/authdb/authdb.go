package authdb

import (
	"fmt"

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

	// InternalUser is the name of the user representing the auth system. By default, it is a cluster admin.
	InternalUser = auth.InternalPrefix + "auth-server"
)

var authConfigIndexes = []*col.Index{}

// AuthConfigCollection returns a postgres collection representing auth configs.
func AuthConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		authConfigCollectionName,
		db,
		listener,
		&auth.OIDCConfig{},
		nil,
	)
}

var roleBindingsIndexes = []*col.Index{}

// RoleBindingCollection returns a postgres collection representing role bindings.
func RoleBindingCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		roleBindingsCollectionName,
		db,
		listener,
		&auth.RoleBinding{},
		roleBindingsIndexes,
	)
}

var membersIndexes = []*col.Index{}

// MembersCollection returns a postgres collection representing members.
func MembersCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		membersCollectionName,
		db,
		listener,
		&auth.Groups{},
		membersIndexes,
	)
}

var groupsIndexes = []*col.Index{}

// GroupsCollection returns a postgres collection representing groups.
func GroupsCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
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
	roleBindings := RoleBindingCollection(nil, nil)
	var binding auth.RoleBinding
	if err := roleBindings.ReadWrite(tx).Get(auth.ClusterRoleBindingKey, &binding); err != nil {
		if col.IsErrNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "getting the cluster role binding")
	}
	binding.Entries[InternalUser] = &auth.Roles{Roles: map[string]bool{auth.ClusterAdminRole: true}}
	return errors.EnsureStack(roleBindings.ReadWrite(tx).Put(auth.ClusterRoleBindingKey, &binding))
}

// ResourceKey generates the key for a resource in the role bindings collection.
func ResourceKey(r *auth.Resource) string {
	if r.Type == auth.ResourceType_CLUSTER {
		return auth.ClusterRoleBindingKey
	}
	return fmt.Sprintf("%s:%s", r.Type, r.Name)
}
