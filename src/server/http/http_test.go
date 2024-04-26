package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestCSRFWrapper(t *testing.T) {
	testData := []struct {
		name        string
		request     func(req *http.Request)
		wantAllowed bool
	}{
		{
			name:        "empty",
			request:     func(req *http.Request) {},
			wantAllowed: true,
		},
		{
			name:        "origin mismatch",
			request:     func(req *http.Request) { req.Header.Add("origin", "http://example.com:1234") },
			wantAllowed: false,
		},
		{
			name:        "referer mismatch",
			request:     func(req *http.Request) { req.Header.Add("referer", "http://example.com:1234/index.html") },
			wantAllowed: false,
		},
		{
			name:        "no host in origin",
			request:     func(req *http.Request) { req.Header.Add("origin", "foo:bar") },
			wantAllowed: false,
		},
		{
			name:        "no host in referer",
			request:     func(req *http.Request) { req.Header.Add("referer", "foo:bar") },
			wantAllowed: false,
		},
		{
			name:        "unparseable origin",
			request:     func(req *http.Request) { req.Header.Add("origin", string([]byte{0x7f})) },
			wantAllowed: false,
		},
		{
			name:        "unparseable referer",
			request:     func(req *http.Request) { req.Header.Add("referer", string([]byte{0x7f})) },
			wantAllowed: false,
		},
		{
			name:        "valid origin",
			request:     func(req *http.Request) { req.Header.Add("origin", "http://example.com") },
			wantAllowed: true,
		},
		{
			name:        "valid referer",
			request:     func(req *http.Request) { req.Header.Add("referer", "http://example.com/index.html") },
			wantAllowed: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var gotCookies []*http.Cookie
			wantCookie := &http.Cookie{
				Name:  "cookie",
				Value: "cookie",
			}
			f := CSRFWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCookies = r.Cookies()
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok")) //nolint:errcheck
			}))

			ctx := pctx.TestContext(t)
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			req = req.WithContext(ctx)
			req.AddCookie(wantCookie)
			test.request(req)
			w := httptest.NewRecorder()
			f(w, req)
			if !test.wantAllowed && len(gotCookies) > 0 {
				t.Errorf("did not want cookies, but got %#v", gotCookies)
			}
			if test.wantAllowed && len(gotCookies) == 0 {
				t.Error("wanted cookies, but got none")
			}
		})
	}
}
