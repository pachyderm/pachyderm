// Command pachhttp runs the Pachyderm HTTP server locally, against your current pachctl context.
package main

import (
	"context"
	"os/signal"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/pachyderm/pachyderm/v2/src/server/restgateway"
	"go.uber.org/zap"
)

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx, cancel := signal.NotifyContext(pctx.Background(""), signals.TerminationSignals...)
	defer cancel()

	pc := &pachctl.Config{Verbose: true}
	c, err := pc.NewOnUserMachine(ctx, false)
	if err != nil {
		log.Exit(ctx, "problem creating pachyderm client", zap.Error(err))
	}
	s := restgateway.New(1660, func(ctx context.Context) *client.APIClient { return c.WithCtx(ctx) })
	log.Info(ctx, "starting server on port 1660")
	if err := s.ListenAndServe(ctx); err != nil {
		log.Exit(ctx, "problem running http server", zap.Error(err))
	}
}
