package archiveserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestHTTP(t *testing.T) {
	testData := []struct {
		name     string
		url      string
		wantCode int
	}{
		{
			name:     "unknown route",
			url:      "http://pachyderm.example.com/what-is-this?",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "health",
			url:      "http://pachyderm.example.com/healthz",
			wantCode: http.StatusOK,
		},
		{
			name:     "empty download",
			url:      "http://pachyderm.example.com/download/AQ.zip",
			wantCode: http.StatusOK,
		},
	}

	s := NewHTTP(0, &client.APIClient{})
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			req := httptest.NewRequest("GET", test.url, nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			s.mux.ServeHTTP(rec, req)
			if got, want := rec.Code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
