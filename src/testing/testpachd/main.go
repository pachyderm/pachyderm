// Command testpachd starts up a local pachd.  It prints its address to stdout when it is ready and
// then runs forever.  It will exit if it encounters an error, or if Control-C is pressed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/cleanup"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

const rootToken = "iamroot"

var (
	logfile      = flag.String("log", "", "if set, log to this file")
	verbose      = flag.Bool("v", false, "if true, show debug level logs (they always end up in logfile, though)")
	pachCtx      = flag.String("context", "testpachd", "if set, setup the named pach context to connect to this server, and switch to it")
	activateAuth = flag.Bool("auth", false, "if true, activate auth")
	useLoki      = flag.Bool("loki", false, "if true, start a loki sidecar and send pachd logs there")
)

func main() {
	var err error
	flag.Parse()

	defer func(done func(error)) {
		done(err)
	}(log.InitBatchLogger(*logfile))
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	var clean cleanup.Cleaner
	defer func() {
		ctx := pctx.Background("cleanup")
		if err = clean.Cleanup(ctx); err != nil {
			log.Error(ctx, "problem cleaning up", zap.Error(err))
		}
	}()

	ctx, cancel := pctx.Interactive()
	defer cancel()

	// Init testpachd options.
	var opts []pachd.TestPachdOption
	if *activateAuth {
		opts = append(opts, pachd.ActivateAuthOption(rootToken))
	}

	if *useLoki {
		var tmpdir string
		if tmpdir, err = os.MkdirTemp("", "testpachd-loki-"); err != nil {
			log.Error(ctx, "problem making tmpdir for loki", zap.Error(err))
			return
		}
		clean.AddCleanup("loki files", func() error {
			return errors.Wrapf(os.RemoveAll(tmpdir), "cleanup loki tmpdir %v", tmpdir)
		})
		l, err := testloki.New(ctx, tmpdir)
		if err != nil {
			log.Error(ctx, "problem starting loki", zap.Error(err))
			return
		}
		clean.AddCleanup("loki", l.Close)
		opt := testloki.WithTestLoki(l)
		opts = append(opts, opt)
	}

	// Build pachd.
	pd, cleanupPachd, err := pachd.BuildTestPachd(ctx, opts...)
	clean.Subsume(cleanupPachd)
	if err != nil {
		// If pachd failed to build, exit now.
		log.Error(ctx, "problem building pachd", zap.Error(err))
		return
	}

	errCh := make(chan error)
	go func(ctx context.Context) {
		defer close(errCh)
		for _, opt := range opts {
			if opt.MutateContext != nil {
				ctx = opt.MutateContext(ctx)
			}
		}
		if err := pd.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}
	}(ctx)

	// Get an RPC client connected to testpachd.
	pachClient, err := pd.PachClient(ctx)
	if err != nil {
		log.Error(ctx, "problem creating pach client", zap.Error(err))
		err = errors.Wrap(err, "create pach client")
		return
	}

	// If the user wants their pachctl config to be updated, do that now.
	if *pachCtx != "" {
		var oldContext string
		if oldContext, err = setupPachctlConfig(ctx, *pachCtx, *activateAuth, pachClient); err != nil {
			log.Error(ctx, "problem reading pachctl config", zap.Error(err))
			return
		}
		if oldContext != "" && oldContext != *pachCtx {
			// Restore the context they were currently pointing at on exit.
			clean.AddCleanupCtx("restore pach context", func(ctx context.Context) error {
				return restorePachctlConfig(ctx, oldContext)
			})
		}
	}

	// Poll pachd until it's ready.
	go func() {
		for {
			if err := pachClient.Health(); err != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Debug(ctx, "pachd not yet healthy, retrying...")
				time.Sleep(time.Second)
				continue
			}
			break
		}
		if *activateAuth {
			log.Info(ctx, "pachd ready; waiting for auth...")
			pd.AwaitAuth(ctx)
		}
		// When ready, print the address on stdout.  This is so that non-Go tests can run
		// testpachd as a subprocess and not have to reimplement health checking / auth
		// readiness checking.  Once a line is printed, it's ready to go.
		fmt.Println(pachClient.GetAddress().Qualified())
		os.Stdout.Close()
	}()

	// wait for pachd to complete
	if err = <-errCh; err != nil {
		log.Error(ctx, "problem running pachd", zap.Error(err))
	}
}

func setupPachctlConfig(ctx context.Context, context string, activateAuth bool, pachClient *client.APIClient) (string, error) {
	pachctl, err := config.Read(true, false)
	if err != nil {
		log.Error(ctx, "problem reading pachctl config", zap.Error(err))
		return "", errors.Wrap(err, "read pachctl config")
	}
	old := pachctl.GetV2().GetActiveContext()
	contexts := pachctl.GetV2().GetContexts()
	c := &config.Context{
		PachdAddress: pachClient.GetAddress().Qualified(),
	}
	if activateAuth {
		c.SessionToken = rootToken
	}
	contexts[*pachCtx] = c
	pachctl.GetV2().ActiveContext = *pachCtx
	if err := pachctl.Write(); err != nil {
		log.Error(ctx, "problem updating pachctl config", zap.Error(err))
		return "", errors.Wrap(err, "write pachctl config")
	}
	log.Info(ctx, "set pachctl context", zap.String("context", context))
	return old, nil
}

func restorePachctlConfig(ctx context.Context, oldContext string) error {
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
}
