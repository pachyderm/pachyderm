package v2_5_0

import (
	"context"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"google.golang.org/protobuf/proto"
)

const (
	defaultProjectName = "default"

	clusterRoleBindingKey       = "CLUSTER:"
	projectRoleBindingKeyPrefix = "PROJECT:"

	allUsersPrincipalKey       = "allClusterUsers"
	pipelinePrincipalKeyPrefix = "pipeline:"

	projectCreatorRole = "projectCreator"
)

func authIsActive(c collection.PostgresReadWriteCollection) bool {
	return !errors.Is(c.Get(clusterRoleBindingKey, &auth.RoleBinding{}), collection.ErrNotFound{})
}

// migrateAuth migrates auth to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do so.
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
	if err := roleBindingsCol.Upsert(clusterRoleBindingKey, clusterRbs, func() error {
		if _, ok := clusterRbs.Entries[allUsersPrincipalKey]; !ok {
			clusterRbs.Entries[allUsersPrincipalKey] = &auth.Roles{Roles: make(map[string]bool)}
		}
		clusterRbs.Entries[allUsersPrincipalKey].Roles[projectCreatorRole] = true
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not update cluster level role bindings")
	}

	// TODO CORE-1048, grant all users the ProjectWriter role for default project
	defaultProjectRbs := &auth.RoleBinding{Entries: make(map[string]*auth.Roles)}
	if err := roleBindingsCol.Upsert(projectRoleBindingKeyPrefix+defaultProjectName, defaultProjectRbs, func() error {
		return nil
	}); err != nil {
		return errors.Wrap(err, "could not update default project's role bindings")
	}

	// Need to extend the character length limit for the subject column because we are adding project name to it.
	if _, err := tx.ExecContext(ctx, `ALTER TABLE auth.auth_tokens ALTER COLUMN subject TYPE varchar(128)`); err != nil {
		return errors.Wrap(err, "could not alter column subject in table auth.auth_tokens from varchar(64) to varchar(128)")
	}

	// Rename pipeline users from "pipeline:<repo>" to "pipeline:default/<repo>"
	if _, err := tx.ExecContext(ctx, `UPDATE auth.auth_tokens SET subject = regexp_replace(subject, '^pipeline:([-a-zA-Z0-9_]+)$', 'pipeline:default/\1') WHERE subject ~ '^pipeline:[^/]+$'`); err != nil {
		return errors.Wrap(err, "could not update auth tokens")
	}
	rb := &auth.RoleBinding{}
	if err := migratePostgreSQLCollection(ctx, tx, "role_bindings", nil, rb, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		newEntries := make(map[string]*auth.Roles)
		for principal, roles := range rb.Entries {
			if strings.HasPrefix(principal, pipelinePrincipalKeyPrefix) && !strings.Contains(principal, "/") {
				principal = pipelinePrincipalKeyPrefix + defaultProjectName + "/" + principal[len(pipelinePrincipalKeyPrefix):]
			}
			newEntries[principal] = roles
		}
		rb.Entries = newEntries
		return oldKey, rb, nil
	}); err != nil {
		return errors.Wrap(err, "could not update pipeline subjects to be aware of default project")
	}

	return nil
}
