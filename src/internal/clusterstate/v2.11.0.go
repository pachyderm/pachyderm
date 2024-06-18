package clusterstate

import (
	v2_11_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.11.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

var state_2_11_0 migrations.State = v2_11_0.Migrate(state_2_10_0)
