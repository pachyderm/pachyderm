package pachd

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/preflight"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

// preflightBuilder builds a preflight-mode pachd instance.
type preflightBuilder struct {
	builder
}

func (pb *preflightBuilder) initServiceEnv(ctx context.Context) error {
	pb.env = serviceenv.InitDBOnlyEnv(ctx, pb.config)
	return nil
}

func (pb *preflightBuilder) setupDB(ctx context.Context) error {
	if err := dbutil.WaitUntilReady(ctx, pb.env.GetDBClient()); err != nil {
		return err
	}
	return nil
}

func (pb *preflightBuilder) testMigrations(ctx context.Context) error {
	return preflight.TestMigrations(ctx, pb.env.GetDBClient())
}

func (pb *preflightBuilder) everythingOK(ctx context.Context) error {
	log.Info(ctx, "all preflight checks OK; it is safe to upgrade this environment to Pachyderm "+version.Version.Canonical())
	return nil
}

func (pb *preflightBuilder) buildAndRun(ctx context.Context) error {
	return pb.apply(ctx, pb.printVersion, pb.tweakResources, pb.initServiceEnv, pb.setupDB, pb.testMigrations, pb.everythingOK)
}

// PreflightMode runs pachd's preflight checks.  It is safe to run these at any time, even for a
// version different than the currently-running full pachd version.
func PreflightMode(ctx context.Context, config *pachconfig.PachdPreflightConfiguration) error {
	log.SetLevel(log.DebugLevel)
	b := &preflightBuilder{
		builder: newBuilder(config, "pachyderm-pachd-preflight"),
	}
	return b.buildAndRun(ctx)
}
