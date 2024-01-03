package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	pachdhttp "github.com/pachyderm/pachyderm/v2/src/server/http"
)

func TestRouting(t *testing.T) {
	testData := []struct {
		name     string
		method   string
		url      string
		body     []byte
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
			url:      "http://pachyderm.example.com/jsonschema/pps_v2/CreatePipelineRequest.schema.json",
			wantCode: http.StatusOK,
		},
		{
			name:     "restgateway InspectCluster",
			method:   "POST",
			url:      "http://pachyderm.example.com/api/admin_v2.API/InspectCluster",
			body:     []byte(`{"clientVersion":{}}`),
			wantCode: http.StatusOK,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)

			s, err := pachdhttp.New(ctx, 0, func(ctx context.Context) *client.APIClient {
				env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
				client := env.PachClient
				return client
			})
			if err != nil {
				t.Fatalf("new pachdhttp server: %v", err)
			}
			log.AddLoggerToHTTPServer(ctx, test.name, s.Server)

			var body io.Reader
			if test.body != nil {
				body = bytes.NewReader(test.body)
			}
			req := httptest.NewRequest(test.method, test.url, body)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			s.Server.Handler.ServeHTTP(rec, req)
			if got, want := rec.Code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
