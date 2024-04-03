// Command testpachd starts up a local pachd.  It prints its address to stdout when it is ready and
// then runs forever.  It will exit if it encounters an error, or if Control-C is pressed.
package main

import (
	"context"
	stderrors "errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	// Build pachd.
	ctx, cancel := pctx.Interactive()
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	// Cleanup pachd on return.
	defer func() {
		if err := eg.Wait(); err != nil {
			log.Error(pctx.Background("testpachd"), "problem running or cleaning up pachd", zap.Error(err))
		}
		cancel()
	}()

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
		defer func() {
			err = stderrors.Join(err, errors.Wrapf(os.RemoveAll(tmpdir), "cleanup loki tmpdir %v", tmpdir))
		}()
		l, err := testloki.New(ctx, tmpdir)
		if err != nil {
			log.Error(ctx, "problem starting loki", zap.Error(err))
			return
		}
		defer l.Close()
		opt := testloki.WithTestLoki(l)
		ctx = testloki.WithLoki(ctx, l)
		opts = append(opts, opt)
	}

	// Build pachd.
	pd, err := pachd.BuildAndRunTestPachd(ctx, eg, opts...)
	if err != nil {
		// If pachd failed to build, exit now.
		log.Error(ctx, "problem building pachd", zap.Error(err))
		return
	}

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
			eg.Go(func() error { return restorePachctlConfig(ctx, oldContext) })
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
	<-ctx.Done()
	cfg, err := config.Read(true, false)
	if err != nil {
		return errors.Wrap(err, "read pachctl config")
	}
	cfg.V2.ActiveContext = oldContext
	if err := cfg.Write(); err != nil {
		return errors.Wrap(err, "write restored context to pachctl config")
	}
	log.Info(pctx.Background("restoring pachctl config"), "restored pachctl config", zap.String("context", oldContext))
	return nil
}
