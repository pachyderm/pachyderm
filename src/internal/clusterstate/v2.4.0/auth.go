package v2_4_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func authIsActive(c collection.PostgresReadWriteCollection) bool {
	return !errors.Is(c.Get(auth.ClusterRoleBindingKey, &auth.RoleBinding{}), collection.ErrNotFound{})
}

// migrateAuth migrates auth to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.
func migrateAuth(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE collections.role_bindings SET key = regexp_replace(key, '^REPO:([-a-zA-Z0-9_]+)$', 'REPO:default/\1') where key ~ '^REPO:([-a-zA-Z0-9_]+)'`); err != nil {
		return errors.Wrap(err, "could not update role bindings")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.role_bindings SET key = regexp_replace(key, '^SPEC_REPO:([-a-zA-Z0-9_]+)$', 'SPEC_REPO:default/\1') where key ~ '^SPEC_REPO:([-a-zA-Z0-9_]+)'`); err != nil {
		return errors.Wrap(err, "could not update role bindings")
	}

	// If auth is already activated, then run the migrations below because they wouldn't have gotten the new role bindings via activation.
	roleBindingsCol := authdb.RoleBindingCollection(nil, nil).ReadWrite(tx)
	if !authIsActive(roleBindingsCol) {
		return nil
	}

	// Grant all users the ProjectCreator role at the cluster level
	clusterRbs := &auth.RoleBinding{Entries: make(map[string]*auth.Roles)}
	if err := roleBindingsCol.Upsert(auth.ClusterRoleBindingKey, clusterRbs, func() error {
		if _, ok := clusterRbs.Entries[auth.AllClusterUsersSubject]; !ok {
			clusterRbs.Entries[auth.AllClusterUsersSubject] = &auth.Roles{Roles: make(map[string]bool)}
		}
		clusterRbs.Entries[auth.AllClusterUsersSubject].Roles[auth.ProjectCreator] = true
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not update cluster level role bindings")
	}

	// Grant all users the ProjectWriter role for default project
	defaultProjectRbs := &auth.RoleBinding{Entries: make(map[string]*auth.Roles)}
	if err := roleBindingsCol.Upsert("PROJECT:default", defaultProjectRbs, func() error {
		if _, ok := defaultProjectRbs.Entries[auth.AllClusterUsersSubject]; !ok {
			defaultProjectRbs.Entries[auth.AllClusterUsersSubject] = &auth.Roles{Roles: make(map[string]bool)}
		}
		defaultProjectRbs.Entries[auth.AllClusterUsersSubject].Roles[auth.ProjectWriter] = true
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not update default project's role bindings")
	}

	return nil
}
