package clusterstate

import (
	v2_10_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.10.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

var state_2_10_0 migrations.State = v2_10_0.Migrate(state_2_8_0)
