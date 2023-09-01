// Package setupenv manages creating pachd.*Envs from pachconfig objects.
package setupenv

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
)

func NewPreflightEnv(config pachconfig.Configuration) (*pachd.PreFlightEnv, error) {
	db, err := openDirectDB(config)
	if err != nil {
		return nil, err
	}
	return &pachd.PreFlightEnv{DB: db}, nil
}
