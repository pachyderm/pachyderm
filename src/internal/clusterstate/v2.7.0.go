package clusterstate

import (
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
)

// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE
var state_2_7_0 migrations.State = state_2_6_0.
	Apply("create cluster defaults", func(ctx context.Context, env migrations.Env) error {
		var cc []col.PostgresCollection
		cc = append(cc, ppsdb.CollectionsV2_7_0()...)
		return col.SetupPostgresCollections(ctx, env.Tx, cc...)
	})
	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
