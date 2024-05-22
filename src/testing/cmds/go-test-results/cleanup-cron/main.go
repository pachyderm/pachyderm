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
	cutoff := time.Now().UTC().Add(-time.Hour * 24 * 365 * 3)
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(gotestresults.PostgresqlHost, 5432),
		dbutil.WithDBName("ci_metrics"),
		dbutil.WithUserPassword(gotestresults.PostgresqlUser, gotestresults.PostgresqlPassword),
	)
	if err != nil {
		return errors.Wrapf(err, "Opening db connection")
	}
	log.Info(ctx, "Deleting logs before cutoff time", zap.String("cutoff", cutoff.String()))
	result, err := db.Query(`
		DELETE FROM public.test_results 
		USING ci_jobs 
		WHERE public.ci_jobs.job_timestamp < $1
		AND ci_jobs.workflow_id=test_results.workflow_id AND ci_jobs.job_id=test_results.job_id AND ci_jobs.job_executor=test_results.job_executor
		RETURNING *`,
		cutoff,
	)
	if err != nil {
		return errors.Wrapf(err, "Deleting old records")
	}
	count := 0
	for result.Next() {
		count++
	}
	log.Info(ctx, "Successfully deleted rows before cutoff. ", zap.Time("cutoff", cutoff), zap.Any("count of deleted rows", count))
	return nil
}
