package pachd

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type prometheusServer struct {
	port uint16
}

// listenAndServe listens until ctx is cancelled; it then gracefully shuts down
// the server, returning once all requests have been handled.
func (ps prometheusServer) listenAndServe(ctx context.Context) error {
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
		// NOTE: using context.Background here means that shutdown will
		// wait until all requests terminate.
		log.Info("terminating S3 server due to cancelled context")
		return errors.EnsureStack(srv.Shutdown(context.Background()))
	case err := <-errCh:
		return err
	}
}
