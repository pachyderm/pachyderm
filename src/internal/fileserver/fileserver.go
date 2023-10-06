// Package fileserver implements a server for downloading PFS files over plain HTTP (i.e. browser downloads).
//
// See: https://www.notion.so/2023-04-03-HTTP-file-and-archive-downloads-cfb56fac16e54957b015070416b09e94
package fileserver

import (
	"context"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/conditionalrequest"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth/httpauth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/timewasted/go-accept-headers"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	//go:embed templates/*
	templateFS embed.FS
	templates  = template.Must(template.New("").Funcs(template.FuncMap{
		"humanizeBytes": func(x int64) string {
			return humanize.Bytes(uint64(x))
		},
		"basename": func(x string) string {
			return path.Base(x)
		},
		"dirname": func(x string) string {
			return path.Dir(x)
		},
		"path": func(info *pfs.FileInfo) string {
			return path.Clean(path.Join(info.GetFile().GetCommit().GetRepo().GetProject().GetName(),
				info.GetFile().GetCommit().GetRepo().GetName(),
				info.GetFile().GetCommit().GetId(),
				info.GetFile().GetPath()))
		},
		"rfc3339": func(x time.Time) string {
			return x.Format(time.RFC3339)
		},
	}).ParseFS(templateFS, "templates/*"))
)

// Server is an http.Handler that can download from and upload to PFS.
type Server struct {
	ClientFactory func(context.Context) *client.APIClient
}

type Request struct {
	PachClient     *client.APIClient
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	RequestID      string
	HTML           bool
}

// ServeHTTP implements http.Handler for the /pfs/ route.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	pachClient := httpauth.ClientWithToken(ctx, s.ClientFactory(ctx), req)
	ctx = pachClient.Ctx()

	ourRequest := &Request{
		Request:        req,
		ResponseWriter: w,
		PachClient:     pachClient,
		RequestID:      log.RequestID(ctx),
		HTML:           acceptsHTML(req),
	}

	switch req.Method {
	case http.MethodHead, http.MethodGet:
		w.Header().Set("vary", constants.ContextTokenKey)
		ourRequest.get(ctx)
		return
	default:
		ourRequest.displayErrorf(ctx, http.StatusMethodNotAllowed, "unknown HTTP method %q", req.Method)
		return
	}
}

func (r *Request) get(ctx context.Context) {
	req := r.Request
	parts := strings.Split(req.URL.Path, "/")
	switch {
	case len(parts) == 5:
		http.Redirect(r.ResponseWriter, r.Request, r.Request.URL.Path+"/", http.StatusMovedPermanently)
		return
	case len(parts) >= 6: // /pfs/project/repo/commit|branch/path/to/some/file
		parts[5] = path.Join(parts[5:]...)
	default:
		// Someday we can have a projects page, project repo page, repo commit page, etc.
		r.displayErrorf(ctx, http.StatusBadRequest,
			//                     0 1   2         3      4               5
			"invalid URL; expecting /pfs/<project>/<repo>/<commit|branch>/<path...>, got %v",
			strings.Join(parts, "/"))
		return
	}
	if got, want := parts[1], "pfs"; got != want {
		r.displayErrorf(ctx, http.StatusInternalServerError, "unexpectedly handling request not for /pfs; got %v want %v", got, want)
		return
	}
	file := &pfs.File{
		Path: parts[5],
	}
	repo := &pfs.Repo{
		Name: parts[3],
		Type: pfs.UserRepoType,
		Project: &pfs.Project{
			Name: parts[2],
		},
	}
	if commitish := parts[4]; uuid.IsUUIDWithoutDashes(commitish) {
		file.Commit = &pfs.Commit{
			Repo: repo,
			Id:   commitish,
		}
		var finished bool
		if commit, err := r.PachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: file.Commit,
		}); err == nil {
			finished = commit.GetFinished() != nil
		}
		if finished {
			// If the commit is finished, the user can cache this content
			// forever.
			r.ResponseWriter.Header().Set("cache-control", "private")
		} else {
			// Open commits can change, so they aren't safe to cache.
			r.ResponseWriter.Header().Set("cache-control", "no-cache")
		}
	} else {
		file.Commit = &pfs.Commit{
			Branch: &pfs.Branch{
				Repo: repo,
				Name: commitish,
			},
		}
		// Branch references are never cacheable; the branch can move at any time.
		r.ResponseWriter.Header().Set("cache-control", "no-cache")
	}
	info, err := r.PachClient.PfsAPIClient.InspectFile(ctx, &pfs.InspectFileRequest{
		File: file,
	})
	if err != nil {
		r.displayGRPCError(ctx, "problem inspecting file", err)
		return
	}

	// Set headers for future conditional requests.
	etag := `"` + hex.EncodeToString(info.GetHash()) + `"`
	lastModified := info.GetCommitted().AsTime()
	if info.GetCommitted() != nil {
		r.ResponseWriter.Header().Set("last-modified", lastModified.In(time.UTC).Format(http.TimeFormat))
	} else {
		lastModified = time.Time{}
	}
	r.ResponseWriter.Header().Set("etag", etag)

	// If a directory, show a directory listing.
	if info.GetFileType() == pfs.FileType_DIR {
		if !strings.HasSuffix(r.Request.URL.Path, "/") {
			http.Redirect(r.ResponseWriter, r.Request, r.Request.URL.Path+"/", http.StatusMovedPermanently)
			return
		}
		r.displayDirectoryListing(ctx, info)
		return
	}

	// Since this is a file, we can accept range requests.
	r.ResponseWriter.Header().Set("accept-ranges", "bytes")

	// Evaluate any HTTP conditional request options.
	status := conditionalrequest.Evaluate(req, &conditionalrequest.ResourceInfo{
		LastModified: lastModified,
		ETag:         etag,
	})
	if status != 0 && status != http.StatusPartialContent {
		// Bail out early; the HTTP precondition failed.
		r.ResponseWriter.WriteHeader(status)
		return
	}

	// If a file, send the file.
	if req.Method == http.MethodHead {
		// If HEAD, they either got 304 Not Modified from Evaluate above, or just want to
		// know the size/modifiation time.
		r.ResponseWriter.Header().Set("content-length", strconv.FormatInt(info.GetSizeBytes(), 10))
		return
	}
	r.sendFile(ctx, info)
}

func (r *Request) displayDirectoryListing(ctx context.Context, info *pfs.FileInfo) {
	ctx, c := pctx.WithCancel(ctx)
	defer c()

	res, err := r.PachClient.PfsAPIClient.ListFile(ctx, &pfs.ListFileRequest{
		File: info.GetFile(),
	})
	if err != nil {
		r.displayGRPCError(ctx, "problem listing directory", err)
	}
	w := r.ResponseWriter
	if r.HTML {
		if err := templates.ExecuteTemplate(w, "directory-listing-header.html", info); err != nil {
			log.Info(ctx, "problem executing directory-listing-header.html template", zap.Error(err))
			fmt.Fprintf(w, "\n\nerror executing template: %v", err)
			return
		}
		defer func() {
			if err := templates.ExecuteTemplate(w, "directory-listing-footer.html", nil); err != nil {
				log.Info(ctx, "problem executing directory-listing-footer.html template", zap.Error(err))
				fmt.Fprintf(w, "\n\nerror executing template: %v", err)
				return

			}
		}()
	}
	for {
		msg, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Info(ctx, "directory listing stream broke unexpectedly", zap.Error(err))
			fmt.Fprintf(w, "\n\nstream broken: %v", err)
			return
		}
		if r.HTML {
			if err := templates.ExecuteTemplate(w, "directory-listing-row.html", msg); err != nil {
				log.Info(ctx, "problem executing directory-listing-row.html template", zap.Error(err))
				fmt.Fprintf(w, "\n\nerror executing template: %v", err)
				return
			}
		} else {
			fmt.Fprintf(w, "%s\t%s\t%v\n", msg.GetFile().GetPath(), msg.GetFile().GetDatum(), msg.GetSizeBytes())
		}
	}
}

// very limited case of range handling, only 'bytes=0-52', no other formats.  the format is
// inclusive; 0-499 returns 500 bytes.
func parseRange(rangeStr string, length int64) (offset, limit int64) {
	if rangeStr == "" || !strings.HasPrefix(rangeStr, "bytes=") {
		return 0, length
	}
	rangeStr = strings.TrimPrefix(rangeStr, "bytes=")
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, length
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, length
	}
	end, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, length
	}
	if end > length {
		end = length - 1
	}
	return start, end - start + 1
}

func (r *Request) sendFile(ctx context.Context, info *pfs.FileInfo) {
	ctx, c := pctx.WithCancel(ctx)
	defer c()

	offset, limit := parseRange(r.Request.Header.Get("range"), info.GetSizeBytes())
	res, err := r.PachClient.PfsAPIClient.GetFile(ctx, &pfs.GetFileRequest{
		File:   info.GetFile(),
		Offset: offset,
	})
	if err != nil {
		r.displayGRPCError(ctx, "problem starting download", err)
		return
	}
	w := r.ResponseWriter
	w.Header().Set("content-length", strconv.FormatInt(limit, 10))
	if limit != info.GetSizeBytes() {
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	for {
		msg, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Info(ctx, "download stream broke unexpectedly", zap.Error(err))
			fmt.Fprintf(w, "\n\nstream broke: %v", err)
			return
		}
		buf := msg.GetValue()
		if len(buf) > int(limit) {
			buf = buf[0:limit]
		}
		n, err := w.Write(buf)
		if err != nil {
			log.Info(ctx, "write error during download", zap.Error(err))
			return
		}
		limit -= int64(n)
		if limit <= 0 {
			return
		}
	}
}

func acceptsHTML(req *http.Request) bool {
	// curl sends "*/*", which is why we prefer text/plain.  browsers explicitly set text/html
	// above */*, so browsers still get HTML.
	got, _ := accept.Negotiate(req.Header.Get("accept"), "text/plain", "text/html")
	return got == "text/html"
}

func (r *Request) displayGRPCError(ctx context.Context, msg string, err error) {
	code := http.StatusInternalServerError
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			code = http.StatusNotFound
		case codes.Unauthenticated:
			code = http.StatusUnauthorized
		case codes.PermissionDenied:
			code = http.StatusForbidden
		}
	}
	r.displayErrorf(ctx, code, "%s: %s", msg, err.Error())
}

func (r *Request) displayErrorf(ctx context.Context, code int, format string, args ...any) {
	w := r.ResponseWriter
	msg := fmt.Sprintf(format, args...)
	var err error
	if r.HTML {
		w.Header().Set("content-type", "text/html")
		w.WriteHeader(code)
		err = templates.ExecuteTemplate(w, "error.html", struct {
			Message   string
			RequestID string
		}{msg, r.RequestID})
	} else {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(code)
		_, err = io.Copy(w, strings.NewReader(msg))
	}
	if err != nil {
		log.Info(ctx, "failed to display error", zap.Error(err), zap.String("message", msg))
	}
}
