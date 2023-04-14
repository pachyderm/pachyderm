// Package archiveserver implements an HTTP server for downloading archives.
//
// See: https://www.notion.so/2023-04-03-HTTP-file-and-archive-downloads-cfb56fac16e54957b015070416b09e94
package archiveserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

type Server struct {
	pachClientFactory func(context.Context) *client.APIClient
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if got, want := req.Method, http.MethodGet; got != want {
		log.Debug(ctx, "invalid method", zap.String("method", got))
		http.Error(w, fmt.Sprintf("method not supported; got %v, want %v", got, want), http.StatusMethodNotAllowed)
		return
	}

	pachClient, err := s.pachClientFromRequest(ctx, req)
	if err != nil {
		log.Info(ctx, "problem extracting auth token from request", zap.Error(err))
		http.Error(w, fmt.Sprintf("problem extracting auth token from request: %v", err), http.StatusBadRequest)
		return
	}

	archive, err := ArchiveFromURL(req.URL)
	if err != nil {
		log.Debug(ctx, "invalid URL", zap.String("url", req.URL.String()))
		http.Error(w, fmt.Sprintf("invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.download(pachClient.Ctx(), w, pachClient, archive); err != nil {
		log.Info(ctx, "problem encountered mid-download", zap.Error(err))
		return
	}
}

func (s *Server) pachClientFromRequest(ctx context.Context, req *http.Request) (*client.APIClient, error) {
	c := s.pachClientFactory(ctx)
	if token := req.URL.Query().Get(constants.ContextTokenKey); token != "" {
		log.Debug(ctx, "using authn-token from URL query", zap.Int("len", len(token)))
		c.SetAuthToken(token)
		return c, nil
	}
	if token := req.Header.Get(constants.ContextTokenKey); token != "" {
		log.Debug(ctx, "using authn-token from HTTP header", zap.Int("len", len(token)))
		c.SetAuthToken(token)
		return c, nil
	}
	return c, nil
}

func (s *Server) download(ctx context.Context, w http.ResponseWriter, pachClient *client.APIClient, req *ArchiveRequest) error {
	fl, ok := w.(http.Flusher)
	if !ok {
		return errors.New("cannot send chunked response")
	}
	w.Header().Add("transfer-encoding", "chunked")
	w.Header().Add("content-disposition", "attachment; filename=download."+string(req.Format))
	w.Header().Add("content-type", req.Format.ContentType())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //nolint:errcheck
	fl.Flush()
	log.Debug(ctx, "archive download done ok")
	return nil
}
