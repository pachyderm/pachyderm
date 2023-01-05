//go:build cgo
// +build cgo

package serviceenv

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
)

func (env *NonblockingServiceEnv) InitDexDB() {
	env.dexDBEg.Go(func() (err error) {
		log.Error(env.Context(), "cannot init dex db when built with CGO_ENABLED=1")
		return errors.New("CGO_ENABLED=1, cannot use dex")
	})
}
