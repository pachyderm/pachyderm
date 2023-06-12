package main

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

const retentionPeriod = time.Hour * 24 * 365 * 3

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

	return nil
}
