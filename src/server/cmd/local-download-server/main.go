// Command local-download-server runs the PFS download server locally, against your current pachctl
// context.
package main

import (
	"context"
	"os/signal"

	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
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
	http := archiveserver.NewHTTP(1659, func(ctx context.Context) *client.APIClient { return c.WithCtx(ctx) })
	log.Info(ctx, "starting server")
	if err := http.ListenAndServe(ctx); err != nil {
		log.Exit(ctx, "problem running http server", zap.Error(err))
	}
}
