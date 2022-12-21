package pachd

import (
	"context"
	"crypto/tls"
	"time"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	pachtls "github.com/pachyderm/pachyderm/v2/src/internal/tls"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"
)

type s3Server struct {
	clientFactory s3.ClientFactory
	port          uint16
}

// listenAndServe listens until ctx is cancelled; it then gracefully shuts down
// the server, returning once all requests have been handled.
func (ss s3Server) listenAndServe(ctx context.Context, shutdownTimeout time.Duration) error {
	ctx = pctx.Child(ctx, "s3", pctx.WithServerID())
	var (
		router = s3.Router(ctx, s3.NewMasterDriver(), ss.clientFactory)
		srv    = s3.Server(ctx, ss.port, router)
		errCh  = make(chan error, 1)
	)
	go func() {
		var (
			certPath, keyPath, err = pachtls.GetCertPaths()
		)
		if err != nil {
			log.Info(ctx, "s3gateway TLS disabled", zap.Error(err))
			errCh <- errors.EnsureStack(srv.ListenAndServe())
		}
		cLoader := pachtls.NewCertLoader(certPath, keyPath, pachtls.CertCheckFrequency)
		// Read TLS cert and key
		if err := cLoader.LoadAndStart(); err != nil {
			errCh <- errors.Wrapf(err, "couldn't load TLS cert for s3gateway: %v", err)
		}
		srv.TLSConfig = &tls.Config{GetCertificate: cLoader.GetCertificate}
		errCh <- errors.EnsureStack(srv.ListenAndServeTLS(certPath, keyPath))
	}()
	select {
	case <-ctx.Done():
		log.Info(ctx, "terminating S3 server due to cancelled context", zap.Error(ctx.Err()))
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return errors.EnsureStack(srv.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
