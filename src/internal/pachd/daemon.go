package pachd

import (
	"context"
	"net/http"
	"os"
	"runtime/pprof"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
)

// Daemon is a Pachyderm daemon.
type Daemon struct {
	// servers
	internal, external *grpcutil.Server
	s3                 *s3Server
	prometheus         *prometheusServer

	// configuration
	criticalServersOnly bool
}

func (d *Daemon) serve(ctx context.Context) (err error) {
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
		eg.Go(maybeIgnoreErrorFunc("S3 Server", !d.criticalServersOnly, func() error { return d.s3.listenAndServe(ctx) }))
	}
	if d.prometheus != nil {
		eg.Go(maybeIgnoreErrorFunc("Prometheus Server", !d.criticalServersOnly, func() error { return d.prometheus.listenAndServe(ctx) }))
	}
	eg.Go(func() error {
		<-ctx.Done() // wait for main context to complete
		var eg = new(errgroup.Group)
		log.Println("terminating; waiting for paused pachd server to gracefully stop")
		eg.Go(func() error { d.external.Server.GracefulStop(); return nil })
		eg.Go(func() error { d.internal.Server.GracefulStop(); return nil })
		err := eg.Wait()
		if err != nil {
			log.Errorf("error waiting for paused pachd server to gracefully stop: %v", err)
		} else {
			log.Println("gRPC server gracefully stopped")
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

func (d Daemon) forGRPCServer(f func(*grpc.Server)) {
	if d.internal != nil {
		f(d.internal.Server)
	}
	if d.external != nil {
		f(d.external.Server)
	}
}
