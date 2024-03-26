// Command testpachd starts up a local pachd.  It prints its address to stdout when it is ready and
// then runs forever.  It will exit if it encounters an error, or if Control-C is pressed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

var (
	logfile = flag.String("log", "", "if set, log to this file")
	verbose = flag.Bool("v", false, "if true, show debug level logs (they always end up in logfile, though)")
	pachCtx = flag.String("context", "testpachd", "if set, setup the named pach context to connect to this server, and switch to it")
)

func main() {
	flag.Parse()
	done := log.InitBatchLogger(*logfile)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	ctx, cancel := pctx.Interactive()
	pd, cleanup, err := pachd.BuildTestPachd(ctx)
	if err != nil {
		log.Error(ctx, "problem building pachd", zap.Error(err))
		if err := cleanup.Cleanup(ctx); err != nil {
			log.Error(ctx, "problem cleaning up after failed run", zap.Error(err))
		}
	}

	errCh := make(chan error)
	go func() {
		if err := pd.Run(ctx); err != nil {
			cancel()
			if !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}
		close(errCh)
	}()

	var exitErr error
	pachClient, err := pd.PachClient(ctx)
	if err != nil {
		log.Error(ctx, "problem creating pach client", zap.Error(err))
		exitErr = errors.Wrap(err, "create pach client")
		cancel()
	}

	if *pachCtx != "" {
		oldContext, err := setupPachctlConfig(ctx, *pachCtx, pachClient)
		if err != nil {
			exitErr = err
			cancel()
		}
		if oldContext != "" && oldContext != *pachCtx {
			cleanup.AddCleanupCtx("restore pach context", func(ctx context.Context) error {
				cfg, err := config.Read(true, false)
				if err != nil {
					return errors.Wrap(err, "read pachctl config")
				}
				cfg.V2.ActiveContext = oldContext
				if err := cfg.Write(); err != nil {
					return errors.Wrap(err, "write restored context to pachctl config")
				}
				log.Info(ctx, "restored pachctl config", zap.String("context", oldContext))
				return nil
			})
		}
	}
	go func() {
		for {
			if err := pachClient.Health(); err != nil {
				log.Debug(ctx, "pachd not yet healthy, retrying...")
				time.Sleep(time.Second)
				continue
			}
			break
		}
		fmt.Println(pachClient.GetAddress().Qualified())
		os.Stdout.Close()
	}()

	<-ctx.Done()
	if err := <-errCh; err != nil {
		log.Error(ctx, "problem running pachd", zap.Error(err))
		exitErr = err
	}
	ctx = pctx.Background("cleanup")
	if err := cleanup.Cleanup(ctx); err != nil {
		log.Error(ctx, "problem cleaning up", zap.Error(err))
	}
	done(exitErr)
}

func setupPachctlConfig(ctx context.Context, context string, pachClient *client.APIClient) (string, error) {
	pachctl, err := config.Read(true, false)
	if err != nil {
		log.Error(ctx, "problem reading pachctl config", zap.Error(err))
		return "", errors.Wrap(err, "read pachctl config")
	}
	old := pachctl.GetV2().GetActiveContext()
	contexts := pachctl.GetV2().GetContexts()
	contexts[*pachCtx] = &config.Context{
		PachdAddress: pachClient.GetAddress().Qualified(),
	}
	pachctl.GetV2().ActiveContext = *pachCtx
	if err := pachctl.Write(); err != nil {
		log.Error(ctx, "problem updating pachctl config", zap.Error(err))
		return "", errors.Wrap(err, "write pachctl config")
	}
	log.Info(ctx, "set pachctl context", zap.String("context", context))
	return old, nil
}
