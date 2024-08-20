package main

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx := pctx.Background("")

	sctx, done := log.SpanContext(ctx, "postgres")
	if err := dockertestenv.EnsureDBEnv(sctx); err != nil {
		done(zap.Error(err))
		log.Exit(ctx, "error", zap.Error(err))
	}
	done()

	sctx, done = log.SpanContext(ctx, "minio")
	if _, err := dockertestenv.NewMinioClient(sctx); err != nil {
		done(zap.Error(err))
		log.Exit(ctx, "error", zap.Error(err))
	}
	done()
}
