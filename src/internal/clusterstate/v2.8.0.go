package clusterstate

import (
	v2_8_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.8.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

var state_2_8_0 migrations.State = v2_8_0.Migrate(state_2_7_0)
