package pfshttp

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/sirupsen/logrus"
)

// Server is an HTTP server that exposes PFS over a browser-compatible REST API.
type Server struct {
	Logger *logrus.Entry
	Client *client.APIClient
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := s.serve(w, req); err != nil {
		status := http.StatusInternalServerError
		var hErr *Error
		if errors.As(err, &hErr) {
			status = hErr.Code
		}
		http.Error(w, err.Error(), status)
		return
	}
}

// Error is an error with an HTTP status code.
type Error struct {
	Code int
	err  error
}

var (
	ErrNotFound = WithCode(http.StatusNotFound, errors.New("not found"))
	ErrUsage    = WithCode(http.StatusBadRequest, errors.New("usage: /pfs/repo@branch/path/to/file"))
)

// Error implements error.
func (err Error) Error() string {
	return err.err.Error()
}

// Unwrap implements errors.Unwrap.
func (err Error) Unwrap() error {
	return err.err
}

// WithCode adds an HTTP status code to an error.
func WithCode(code int, err error) error {
	if code == http.StatusOK && err == nil {
		return nil
	}
	return &Error{
		Code: code,
		err:  err,
	}
}

// serve handles an HTTP request with the option of returning an error.
func (s *Server) serve(w http.ResponseWriter, req *http.Request) error {
	if req.URL == nil {
		return WithCode(http.StatusBadRequest, errors.New("no url in request"))
	}

	l := s.Logger.WithField("method", req.Method).WithField("path", req.URL.Path)
	pathParts := strings.Split(req.URL.Path, "/")
	if len(pathParts) < 2 {
		http.Redirect(w, req, "/pfs", http.StatusFound)
		return nil
	}

	// pathParts[0] is the empty string.
	switch pathParts[1] {
	case "healthz":
		c := s.Client.WithCtx(req.Context())
		if _, err := c.Version(); err != nil {
			l.Info("served 'unhealthy' to pfshttp health check")
			return WithCode(http.StatusInternalServerError, err)
		}
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		l.Debug("served 'healthy' to pfshttp health check")
		return nil
	case "pfs":
		return s.servePFS(w, req, pathParts[2:])
	}
	return ErrNotFound
}

// servePFS serves a file or directory out of PFS.
func (s *Server) servePFS(w http.ResponseWriter, req *http.Request, parts []string) error {
	if len(parts) == 0 {
		return ErrUsage
	}
	l := s.Logger.WithField("method", req.Method).WithField("path", strings.Join(parts, "/"))
	commit, err := cmdutil.ParseCommit(parts[0])
	if err != nil {
		l.WithError(err).Info("failed to serve pfs file")
		return WithCode(http.StatusBadRequest, err)
	}
	path := strings.Join(parts[1:], "/")
	file := &pfs.File{
		Commit: commit,
		Path:   path,
	}

	ctx := req.Context()
	c := s.Client.WithCtx(ctx)

	if t := req.Header.Get("authorization"); t != "" {
		c.SetAuthToken(t)
		ctx = c.Ctx()
		if who, err := c.WhoAmI(ctx, &auth.WhoAmIRequest{}); err == nil {
			l = l.WithField("username", who.GetUsername())
		}
	}

	lfc, err := c.PfsAPIClient.ListFile(ctx, &pfs.ListFileRequest{File: file})
	if err != nil {
		l.WithError(err).Info("failed to serve pfs file (list file)")
		return err
	}
	fi, err := lfc.Recv()
	if err != nil {
		l.WithError(err).Info("failed to serve pfs file (read first entry)")
		return err
	}
	if t := fi.FileType; t == pfs.FileType_DIR {
		// Serve directory listing.
		return nil
	}

	gfc, err := c.PfsAPIClient.GetFile(ctx, &pfs.GetFileRequest{File: file})
	if err != nil {
		l.WithError(err).Info("failed to serve pfs file (get file)")
		if strings.Contains(err.Error(), "cannot get directory") {
			// This error is not specific enough to determine if this is a directory or
			// a non-regular file.

			return WithCode(http.StatusBadRequest, err)
		}
		return err
	}
	var n int
	for m, err := gfc.Recv(); err != io.EOF; m, err = gfc.Recv() {
		if err != nil {
			l.WithError(err).Infof("failed to serve pfs file after %d bytes (read error)", n)
			return err
		}
		if _, err := w.Write(m.Value); err != nil {
			l.WithError(err).Infof("failed to serve pfs file after %d bytes (write error)", n)
			return err
		}
		n += len(m.Value)
	}
	l.Info("served pfs file ok")
	return nil
}
