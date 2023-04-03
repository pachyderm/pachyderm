package v2_6_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"

	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
)

func authIsActive(c collection.PostgresReadWriteCollection) bool {
	return !errors.Is(c.Get("CLUSTER:", &auth.RoleBinding{}), collection.ErrNotFound{})
}

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Grant all users ProjectWriter role for the default project", func(ctx context.Context, env migrations.Env) error {
			roleBindingsCol := authdb.RoleBindingCollection(nil, nil).ReadWrite(env.Tx)
			if !authIsActive(roleBindingsCol) {
				return nil
			}
			rb := &auth.RoleBinding{}
			if err := roleBindingsCol.Upsert("PROJECT:default", rb, func() error {
				if rb.Entries == nil {
					rb.Entries = make(map[string]*auth.Roles)
				}
				if _, ok := rb.Entries[auth.AllClusterUsersSubject]; !ok {
					rb.Entries[auth.AllClusterUsersSubject] = &auth.Roles{Roles: make(map[string]bool)}
				}
				rb.Entries[auth.AllClusterUsersSubject].Roles[auth.ProjectWriterRole] = true
				return nil
			}); err != nil {
				return errors.Wrap(err, "could not update default project role bindings for allClusterUsers")
			}
			return nil
		}).
		Apply("Add commit_provenance table", func(ctx context.Context, env migrations.Env) error {
			return pfsdb.SetupCommitProvenanceV0(ctx, env.Tx)
		}).
		Apply("Remove Alias Commits", func(ctx context.Context, env migrations.Env) error {
			return removeAliasCommits(ctx, env.Tx)
		}).
		Apply("Remove branch from the Commit key", func(ctx context.Context, env migrations.Env) error {
			if err := branchlessCommitsPFS(ctx, env.Tx); err != nil {
				return err
			}
			return branchlessCommitsPPS(ctx, env.Tx)
		}).
		Apply("Add foreign key constraints on pfs.commits.commit_id -> collections.commits.key", func(ctx context.Context, env migrations.Env) error {
			return pfsdb.SetupCommitProvenanceV01(ctx, env.Tx)
		})

	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
}
