package pachd

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/preflight"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

type PreFlightEnv struct {
	// DB should be a direct database connection.
	DB *sqlx.DB
}

// PreFlight is a minimal pachd for running preflight checks.
type PreFlight struct {
	base
	env    PreFlightEnv
	config pachconfig.PachdPreflightConfiguration
}

func NewPreflight(env PreFlightEnv, config pachconfig.PachdPreflightConfiguration) *PreFlight {
	pf := &PreFlight{env: env, config: config}
	pf.addSetup("print version", pf.printVersion)
	pf.addSetup("await DB", pf.awaitDB)
	pf.addSetup("test migrations", pf.testMigrations)
	pf.addSetup("", pf.everythingOK)
	return pf
}

func (pf *PreFlight) awaitDB(ctx context.Context) error {
	if err := dbutil.WaitUntilReady(ctx, pf.env.DB); err != nil {
		return err
	}
	return nil
}

func (pf *PreFlight) testMigrations(ctx context.Context) error {
	return preflight.TestMigrations(ctx, pf.env.DB)
}

func (pf *PreFlight) everythingOK(ctx context.Context) error {
	log.Info(ctx, "all preflight checks OK; it is safe to upgrade this environment to Pachyderm "+version.Version.Canonical())
	return nil
}
