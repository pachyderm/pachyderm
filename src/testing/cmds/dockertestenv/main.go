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
	if err := dockertestenv.EnsureDBEnv(ctx); err != nil {
		log.Exit(ctx, "error", zap.Error(err))
	}
}
