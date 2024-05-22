// Package setupenv manages creating pachd.*Envs from pachconfig objects.
package setupenv

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
)

func NewPreflightEnv(config pachconfig.PachdPreflightConfiguration) (*pachd.PreFlightEnv, error) {
	db, err := openDirectDB(config.PostgresConfiguration)
	if err != nil {
		return nil, err
	}
	return &pachd.PreFlightEnv{DB: db}, nil
}
