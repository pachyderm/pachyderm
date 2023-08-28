package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestRouting(t *testing.T) {
	testData := []struct {
		name     string
		method   string
		url      string
		wantCode int
	}{
		{
			name:     "not found",
			method:   "GET",
			url:      "http://pachyderm.example.com/",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "health",
			method:   "GET",
			url:      "http://pachyderm.example.com/healthz",
			wantCode: http.StatusOK,
		},
		{
			name:     "CreatePipelineRequest JSON schema",
			method:   "GET",
			url:      "http://pachyderm.exmaple.com/jsonschema/CreatePipelineRequest.schema.json",
			wantCode: http.StatusOK,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			s := New(0, func(ctx context.Context) *client.APIClient { return nil })
			log.AddLoggerToHTTPServer(ctx, test.name, s.server)

			req := httptest.NewRequest(test.method, test.url, nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			s.server.Handler.ServeHTTP(rec, req)
			if got, want := rec.Code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}

func TestCSRFWrapper(t *testing.T) {
	testData := []struct {
		name     string
		request  func(req *http.Request)
		wantCode int
	}{
		{
			name:     "empty",
			request:  func(req *http.Request) {},
			wantCode: http.StatusOK,
		},
		{
			name:     "origin mismatch",
			request:  func(req *http.Request) { req.Header.Add("origin", "http://example.com:1234") },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "referer mismatch",
			request:  func(req *http.Request) { req.Header.Add("referer", "http://example.com:1234/index.html") },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "no host in origin",
			request:  func(req *http.Request) { req.Header.Add("origin", "foo:bar") },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "no host in referer",
			request:  func(req *http.Request) { req.Header.Add("referer", "foo:bar") },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "unparseable origin",
			request:  func(req *http.Request) { req.Header.Add("origin", string([]byte{0x7f})) },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "unparseable referer",
			request:  func(req *http.Request) { req.Header.Add("referer", string([]byte{0x7f})) },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "valid origin",
			request:  func(req *http.Request) { req.Header.Add("origin", "http://example.com") },
			wantCode: http.StatusOK,
		},
		{
			name:     "valid referer",
			request:  func(req *http.Request) { req.Header.Add("referer", "http://example.com/index.html") },
			wantCode: http.StatusOK,
		},
	}

	f := CSRFWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	}))
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			req = req.WithContext(ctx)
			test.request(req)
			w := httptest.NewRecorder()
			f(w, req)
			if got, want := w.Code, test.wantCode; got != want {
				t.Errorf("code:\n  got: %v\n want: %v", got, want)
			}
		})
	}

}
