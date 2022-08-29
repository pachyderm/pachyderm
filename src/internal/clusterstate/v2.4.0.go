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
		var defaultProject = &pfs.ProjectInfo{
			Project: &pfs.Project{
				Name: pfs.DefaultProjectName,
			},
		}
		if err := pfsdb.Projects(nil, nil).ReadWrite(env.Tx).Create(pfs.DefaultProjectName, defaultProject); err != nil {
			return errors.Wrap(err, "could not create default project")
		}
		return nil
	})
	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
