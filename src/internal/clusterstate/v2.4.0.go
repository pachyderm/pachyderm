package clusterstate

import (
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

var state_2_4_0 migrations.State = state_2_3_0.
	Apply("Add projects collection", func(ctx context.Context, env migrations.Env) error {
		return col.SetupPostgresCollections(ctx, env.Tx, pfsdb.CollectionsV2_4_0()...)
	}).
	Apply("Add default project", func(ctx context.Context, env migrations.Env) error {
		col := pfsdb.Projects(nil, nil).ReadWrite(env.Tx)
		if err := col.Create(pfs.DefaultProject, &pfs.Project{Name: pfs.DefaultProject}); err != nil {
			return errors.Errorf("could not create default project: %w", err)
		}
		return nil
	})
	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
