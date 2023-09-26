// Package pachd implements the Pachyderm dæmon and its various modes.
//
// Callers need only provide a context and a configuration.
//
// # Adding a new mode
//
// The idea is that there is a builder for each mode; each builder is
// responsible for building and running an instance of the single type daemon
// representing a pachd instance.
//
// To add a new mode one will at least add a new builder; one may also need to
// add new members to daemon.  Daemon should contain only those members needed
// at run time for any mode; other, transient, values should be members of the
// pertinent builder.
package pachd

import (
	"context"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	pachhttp "github.com/pachyderm/pachyderm/v2/src/server/http"
)

// daemon is a Pachyderm daemon.
type daemon struct {
	// servers
	internal, external *grpcutil.Server
	s3                 *s3Server
	prometheus         *prometheusServer
	pachhttp           *pachhttp.Server

	// configuration
	criticalServersOnly bool
}

func (d *daemon) serve(ctx context.Context) (err error) {
	eg, ctx := errgroup.WithContext(pctx.Child(ctx, "serve"))
	go log.WatchDroppedLogs(ctx, time.Minute)
	defer func() {
		if err != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()
	if d.external != nil {
		eg.Go(maybeIgnoreErrorFunc(ctx, "External Pachd GRPC Server", true, func() error {
			return d.external.Wait()
		}))
	}
	if d.internal != nil {
		eg.Go(maybeIgnoreErrorFunc(ctx, "Internal Pachd GRPC Server", true, func() error {
			return d.internal.Wait()
		}))
	}
	if d.s3 != nil {
		eg.Go(maybeIgnoreErrorFunc(ctx, "S3 Server", !d.criticalServersOnly, func() error { return d.s3.listenAndServe(ctx, 10*time.Second) }))
	}
	if d.prometheus != nil {
		eg.Go(maybeIgnoreErrorFunc(ctx, "Prometheus Server", !d.criticalServersOnly, func() error { return d.prometheus.listenAndServe(ctx, 10*time.Second) }))
	}
	if d.pachhttp != nil {
		eg.Go(maybeIgnoreErrorFunc(ctx, "PachHTTP Server", true, func() error { return d.pachhttp.ListenAndServe(ctx) }))
	}
	eg.Go(func() error {
		<-ctx.Done() // wait for main context to complete
		var (
			eg     = new(errgroup.Group)
			doneCh = make(chan struct{})
		)
		log.Info(ctx, "terminating; waiting for pachd server to gracefully stop")
		if d.external != nil {
			eg.Go(func() error { d.external.Server.GracefulStop(); return nil })
		}
		if d.internal != nil {
			eg.Go(func() error { d.internal.Server.GracefulStop(); return nil })
		}
		go func() {
			select {
			case <-time.After(10 * time.Second):
				log.Info(ctx, "pachd did not gracefully stop")
				if d.external != nil {
					d.external.Server.Stop()
					log.Info(ctx, "stopped external server")
				}
				if d.internal != nil {
					d.internal.Server.Stop()
					log.Info(ctx, "stopped internal server")
				}
			case <-doneCh:
				return
			}
		}()
		err := eg.Wait()
		if err != nil {
			log.Error(ctx, "error waiting for servers to gracefully stop", zap.Error(err))
		} else {
			log.Info(ctx, "gRPC server stopped")
			close(doneCh)
		}
		return errors.EnsureStack(err)
	})
	return errors.EnsureStack(eg.Wait())
}

// maybeIgnoreErrorFunc returns a function that runs f; if f returns an HTTP server
// closed error it returns nil; if required is false it returns nil; otherwise
// it returns whatever f returns.
func maybeIgnoreErrorFunc(ctx context.Context, name string, required bool, f func() error) func() error {
	return func() error {
		if err := f(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			if !required {
				log.Error(ctx, "error setting up and/or running server", zap.String("server", name))
				return nil
			}
			return errors.Wrapf(err, "error setting up and/or running %v (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers)", name)
		}
		return nil
	}
}

func (d daemon) forGRPCServer(f func(*grpc.Server)) {
	if d.internal != nil {
		f(d.internal.Server)
	}
	if d.external != nil {
		f(d.external.Server)
	}
}
