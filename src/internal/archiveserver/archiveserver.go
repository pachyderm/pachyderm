// Package archiveserver implements an HTTP server for downloading archives.
//
// See: https://www.notion.so/2023-04-03-HTTP-file-and-archive-downloads-cfb56fac16e54957b015070416b09e94
package archiveserver

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// Server is an http.Handler that can download archives of PFS files.
type Server struct {
	pachClientFactory func(context.Context) *client.APIClient
}

// ServeHTTP implements http.Handler for the /download/ route.
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
	// For now we make no effort to be flexible about the output format.  downloadZip would need
	// to be refactored, because 99% of it is not related to the ZIP format.
	if got, want := archive.Format, ArchiveFormatZip; got != want {
		log.Debug(ctx, "unsupported archive format requested", zap.String("format", string(got)))
		http.Error(w, fmt.Sprintf("unknown archive format %v requested; only .zip is supported", string(got)), http.StatusBadRequest)
		return
	}
	if err := s.downloadZip(pachClient.Ctx(), w, pachClient, archive); err != nil {
		log.Info(ctx, "problem encountered mid-download", zap.Error(err))
		return
	}
	log.Info(ctx, "archive download ok")
}

// pachClientFromRequest uses the server's pachClientFactory to create a pach client with the
// authentication material in the request.
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

// writeFlusher is an io.Writer that flushes the http.Fluher after every write.  It's intended to
// back a bufio.Writer to implement reasonable-sized response chunking.
type writeFlusher struct {
	io.Writer
	http.Flusher
	report func(int)
}

// Write implements io.Writer.
func (w *writeFlusher) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err != nil {
		return n, err //nolint:wrapcheck
	}
	w.report(len(p))
	w.Flush()
	return n, nil
}

func (s *Server) downloadZip(ctx context.Context, rw http.ResponseWriter, pachClient *client.APIClient, req *ArchiveRequest) (retErr error) {
	ctx, done := log.SpanContext(ctx, "downloadZip")
	defer done(log.Errorp(&retErr))

	// Make sure we don't have to buffer the entire response; this should always be ok.
	fl, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "cannot send chunked response; refusing to buffer", http.StatusInternalServerError)
		return errors.New("cannot send chunked response")
	}
	wf := &writeFlusher{
		Writer:  rw,
		Flusher: fl,
		report: func(i int) {
			meters.Inc(ctx, "archive_download_tx_bytes", i)
		},
	}
	// Create a ZIP writer backed by a (chunked) buffer.
	bw := bufio.NewWriterSize(wf, units.MB) // Send an HTTP chunk this often.
	zw := zip.NewWriter(bw)

	// Setup headers for a download based on the current time.
	now := time.Now()
	destPath := fmt.Sprintf("pachyderm-download-%s.%s", now.Format(time.RFC3339), req.Format)
	rw.Header().Add("transfer-encoding", "chunked")
	rw.Header().Add("content-disposition", "attachment; filename="+destPath)
	rw.Header().Add("content-type", req.Format.ContentType())

	// Send OK; even though much can still fail.
	rw.WriteHeader(http.StatusOK)

	// Iterate over every requested path, and copy it into the output ZIP.
	pathErr := req.ForEachPath(func(path string) (retErr error) {
		ctx, done := log.SpanContext(ctx, fmt.Sprintf("downloadPath(%v)", path))
		defer done(log.Errorp(&retErr))

		// Decode the request into a pfs.File.
		file, err := DecodeV1Path(path)
		if err != nil {
			return errors.Wrapf(err, "path %v: decode", path)
		}

		// Ask pachyderm for this file (or directory).
		tr, err := pachClient.WithCtx(ctx).GetFileTAR(file.Commit, file.Path)
		if err != nil {
			return errors.Wrapf(err, "path %v: start GetFileTAR", path)
		}
		defer multierr.AppendInvoke(&retErr, multierr.Close(tr))

		// Decode the TAR from pachd.
		r := tar.NewReader(tr)
		for {
			// Iterate over the TAR headers.
			h, err := r.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// We are done with this path.
					return nil
				}
				return errors.Wrapf(err, "path %v: read TAR header", path)
			}

			// Skip directories, as they do not need to be in the resulting ZIP.
			if h.Typeflag == tar.TypeDir {
				continue
			}

			// Show what files we're putting in the archive, in case things go awry.
			log.Debug(ctx, "got file", zap.String("file", h.Name), zap.String("for_path", path))

			// Create a path in the archive <project>/<repo>/<branch/commit>/<actual
			// path, including directories>.
			zp := filepath.Join(file.Commit.Repo.Project.Name, file.Commit.Repo.Name, file.Commit.Branch.Name, h.Name)
			w, err := zw.CreateHeader(&zip.FileHeader{
				Name:     zp,
				Method:   zip.Deflate,
				Modified: now,
			})
			if err != nil {
				return errors.Wrapf(err, "create zip path %v (for %v in %v)", zp, h.Name, path)
			}
			// Copy the data for this file into the zip.
			n, err := io.Copy(w, r)
			if err != nil {
				return errors.Wrapf(err, "write data for zip path %v (for %v in %v)", zp, h.Name, path)
			}
			meters.Inc(ctx, "archive_download_added_bytes", n)
		}
	})
	if pathErr != nil {
		// If there is an error generated by a per-file callback, try writing it to a file
		// called @error.txt and then completing the ZIP normally.  We'll return this error
		// after flushing the ZIP, so that logs indiciate an error, but the user will have a
		// valid partial ZIP to look at.
		//
		// We choose the name @error.txt because PFS cannot contain a file called
		// @error.txt, but could easily contain a file called error.txt.
		w, zerr := zw.CreateHeader(&zip.FileHeader{
			Name:     "@error.txt",
			Method:   zip.Store, // So the actual bytes of the error appear on the wire.
			Modified: now,
		})
		if zerr != nil {
			// Now we have the exciting situation of an error while handling the error.
			// Bail out with both errors.
			fmt.Fprintf(bw, "\n\ncreate @error.txt: %v\n\ncaused by: %v\n", zerr, pathErr)
			bw.Flush()
			return errors.Errorf("ForEachPath: create @error.txt: %v; caused by %v", zerr, pathErr)
		}
		if _, werr := fmt.Fprintf(w, "%v\n", pathErr); werr != nil {
			// See above; error handling the error, bail out.
			fmt.Fprintf(bw, "\n\nwrite @error.txt: %v\n\ncaused by: %v\n", werr, pathErr)
			bw.Flush()
			return errors.Errorf("ForEachPath: write @error.txt: %v; caused by %v", werr, pathErr)
		}
		// Make a note in the logs that the user did see this error if they look hard
		// enough.
		pathErr = errors.Wrap(pathErr, "(reported via @error.txt)")
	}

	// Finish the ZIP file.
	if err := zw.Close(); err != nil {
		return errors.Wrapf(err, "zip.Writer.Close()")
	}

	// Flush any data in the buffered writer.
	if err := bw.Flush(); err != nil {
		return errors.Wrapf(err, "bufio.Writer.Flush()")
	}

	// Write the last chunk.
	fl.Flush()

	// If there was an error while getting the files, report that to the caller.
	if pathErr != nil {
		return errors.Wrap(pathErr, "ForEachPath")
	}

	// Hey, it worked!
	return nil
}
