package clusterstate

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
)

var state_2_7_0 migrations.State = v2_7_0.Migrate(state_2_6_0)

// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE
