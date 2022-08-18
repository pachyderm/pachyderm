package pachd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type prometheusServer struct {
	port uint16
}

// listenAndServe listens until ctx is cancelled; it then gracefully shuts down
// the server, returning once all requests have been handled.
func (ps prometheusServer) listenAndServe(ctx context.Context, shutdownTimeout time.Duration) error {
	var (
		mux = http.NewServeMux()
		srv = http.Server{
			Addr:        fmt.Sprintf(":%v", ps.port),
			Handler:     mux,
			BaseContext: func(net.Listener) context.Context { return ctx },
		}
		errCh = make(chan error, 1)
	)
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Info("terminating Prometheus server due to cancelled context")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return errors.EnsureStack(srv.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
