package main

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	gotestresults "github.com/pachyderm/pachyderm/v2/src/testing/cmds/go-test-results"
	"go.uber.org/zap"
)

var retentionPeriod = time.Now().UTC().Add(-time.Minute * 10) //-time.Hour * 24 * 365 * 3)

// This is built and runs in a pachyderm cron pipeline to clean up stale data
func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("")
	err := run(ctx)
	if err != nil {
		log.Exit(ctx, "Error during SQL cleanup cron job", zap.Error(err))
	}
}

func run(ctx context.Context) error {
	log.Info(ctx, "Running DB Cleanup")
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(gotestresults.PostgresqlHost, 5432),
		dbutil.WithDBName("ci_metrics"),
		dbutil.WithUserPassword(gotestresults.PostgresqlUser, gotestresults.PostgresqlPassword),
	)
	if err != nil {
		return errors.Wrapf(err, "Opening db connection")
	}
	log.Info(ctx, "Deleting logs before cutoff time", zap.String("cutoff time", retentionPeriod.String()))
	_, err = db.Exec(`
		DELETE FROM public.test_results 
		USING ci_jobs 
		WHERE public.ci_jobs.job_timestamp < $1
		AND ci_jobs.workflow_id=test_results.workflow_id AND ci_jobs.job_id=test_results.job_id AND ci_jobs.job_executor=test_results.job_executor`,
		retentionPeriod,
	)
	if err != nil {
		return errors.Wrapf(err, "Deleting old records")
	}
	return nil
}
