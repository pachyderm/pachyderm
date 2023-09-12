package clusterstate

import (
	v2_8_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.8.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

var state_2_8_0 migrations.State = v2_8_0.Migrate(state_2_7_0)

// todo(fahad): this state will be be merged with state_2_8_0 once the call sites have been updated after the commits migration.
var State_2_8_0_temp migrations.State = v2_8_0.PostMigrate(state_2_8_0)
