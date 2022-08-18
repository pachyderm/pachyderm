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

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
)

// daemon is a Pachyderm daemon.
type daemon struct {
	// servers
	internal, external *grpcutil.Server
	s3                 *s3Server
	prometheus         *prometheusServer

	// configuration
	criticalServersOnly bool
}

func (d *daemon) serve(ctx context.Context) (err error) {
	eg, ctx := errgroup.WithContext(ctx)
	defer func() {
		if err != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()
	if d.external != nil {
		eg.Go(maybeIgnoreErrorFunc("External Pachd GRPC Server", true, func() error {
			return d.external.Wait()
		}))
	}
	if d.internal != nil {
		eg.Go(maybeIgnoreErrorFunc("Internal Pachd GRPC Server", true, func() error {
			return d.internal.Wait()
		}))
	}
	if d.s3 != nil {
		eg.Go(maybeIgnoreErrorFunc("S3 Server", !d.criticalServersOnly, func() error { return d.s3.listenAndServe(ctx, 10*time.Second) }))
	}
	if d.prometheus != nil {
		eg.Go(maybeIgnoreErrorFunc("Prometheus Server", !d.criticalServersOnly, func() error { return d.prometheus.listenAndServe(ctx, 10*time.Second) }))
	}
	eg.Go(func() error {
		<-ctx.Done() // wait for main context to complete
		var (
			eg     = new(errgroup.Group)
			doneCh = make(chan struct{})
		)
		log.Println("terminating; waiting for pachd server to gracefully stop")
		if d.external != nil {
			eg.Go(func() error { d.external.Server.GracefulStop(); return nil })
		}
		if d.internal != nil {
			eg.Go(func() error { d.internal.Server.GracefulStop(); return nil })
		}
		go func() {
			select {
			case <-time.After(10 * time.Second):
				log.Println("pachd did not gracefully stop")
				if d.external != nil {
					d.external.Server.Stop()
					log.Println("stopped external server")
				}
				if d.internal != nil {
					d.internal.Server.Stop()
					log.Println("stopped internal server")
				}
			case <-doneCh:
				return
			}
		}()
		err := eg.Wait()
		if err != nil {
			log.Errorf("error waiting for paused pachd server to gracefully stop: %v", err)
		} else {
			log.Println("gRPC server stopped")
			close(doneCh)
		}
		return errors.EnsureStack(err)
	})
	return errors.EnsureStack(eg.Wait())
}

// maybeIgnoreErrorFunc returns a function that runs f; if f returns an HTTP server
// closed error it returns nil; if required is false it returns nil; otherwise
// it returns whatever f returns.
func maybeIgnoreErrorFunc(name string, required bool, f func() error) func() error {
	return func() error {
		if err := f(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			if !required {
				log.Errorf("error setting up and/or running %v: %v", name, err)
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
