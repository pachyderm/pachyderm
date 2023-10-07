package conditionalrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConditionalRequest(t *testing.T) {
	testData := []struct {
		name    string
		method  string // defaults to GET
		info    *ResourceInfo
		headers http.Header
		want    int
	}{
		{
			name: "no headers",
			want: 0,
		},
		{
			name: "if-modified; not modified",
			headers: http.Header{
				"If-Modified-Since": {"Fri, 07 Oct 2023 00:00:00 GMT"},
			},
			want: http.StatusNotModified,
		},
		{
			name: "if-modified; no modification time known",
			info: &ResourceInfo{},
			headers: http.Header{
				"If-Modified-Since": {"Fri, 07 Oct 2023 00:00:00 GMT"},
			},
			want: 0,
		},
		{
			name: "if-modified; modified",
			headers: http.Header{
				"If-Modified-Since": {"Sun, 01 Dec 2019 00:00:00 GMT"},
			},
			want: 0,
		},
		{
			name: "if-modified; unparseable time",
			headers: http.Header{
				"If-Modified-Since": {"foo bar"},
			},
			want: 0,
		},
		{
			name: "if-unmodified; unmodified",
			headers: http.Header{
				"If-Unmodified-Since": {"Fri, 07 Oct 2023 00:00:00 GMT"},
			},
			want: 0,
		},
		{
			name: "if-unmodified; no modification time known",
			info: &ResourceInfo{},
			headers: http.Header{
				"If-Unmodified-Since": {"Fri, 07 Oct 2023 00:00:00 GMT"},
			},
			want: 0,
		},
		{
			name: "if-unmodified; modified",
			headers: http.Header{
				"If-Unmodified-Since": {"Sun, 01 Dec 2019 00:00:00 GMT"},
			},
			want: http.StatusPreconditionFailed,
		},
		{
			name: "if-unmodified; unparseable time",
			headers: http.Header{
				"If-Unmodified-Since": {"foo bar"},
			},
			want: 0,
		},
		{
			name: "if-match; no match",
			headers: http.Header{
				"If-Match": {"foobar"},
			},
			want: http.StatusPreconditionFailed,
		},
		{
			name: "if-match; match",
			headers: http.Header{
				"If-Match": {`"etag"`},
			},
			want: 0,
		},
		{
			name: "if-match; match *",
			headers: http.Header{
				"If-Match": {`*`},
			},
			want: 0,
		},
		{
			name: "if-none-match; no match",
			headers: http.Header{
				"If-None-Match": {`"foobar"`},
			},
			want: 0,
		},
		{
			name: "if-none-match; match",
			headers: http.Header{
				"If-None-Match": {`"etag"`},
			},
			want: http.StatusNotModified,
		},
		{
			name:   "if-none-match; match; HEAD",
			method: http.MethodHead,
			headers: http.Header{
				"If-None-Match": {`"etag"`},
			},
			want: http.StatusNotModified,
		},
		{
			name:   "if-none-match; match; PUT",
			method: http.MethodPut,
			headers: http.Header{
				"If-None-Match": {`"etag"`},
			},
			want: http.StatusPreconditionFailed,
		},
		{
			name: "if-none-match; match *",
			headers: http.Header{
				"If-None-Match": {`*`},
			},
			want: http.StatusNotModified,
		},
		{
			name: "if-range; match",
			headers: http.Header{
				"Range":    {"bytes=0-0"},
				"If-Range": {`"etag"`},
			},
			want: http.StatusPartialContent,
		},
		{
			name: "if-range; no match",
			headers: http.Header{
				"Range":    {"bytes=0-0"},
				"If-Range": {`"foo"`},
			},
			want: 0,
		},
		{
			name: "if-range; matching date",
			headers: http.Header{
				"Range":    {"bytes=0-0"},
				"If-Range": {`Wed, 01 Jan 2020 00:00:00 GMT`},
			},
			want: http.StatusPartialContent,
		},
		{
			name: "if-range; non-matching date",
			headers: http.Header{
				"Range":    {"bytes=0-0"},
				"If-Range": {`Wed, 01 Jan 2020 00:00:01 GMT`},
			},
			want: 0,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.method == "" {
				test.method = http.MethodGet
			}
			req := httptest.NewRequest(test.method, "/", nil)
			for k, vs := range test.headers {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			info := test.info
			if info == nil {
				info = &ResourceInfo{
					LastModified: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					ETag:         `"etag"`,
				}
			}
			got := Evaluate(req, info)
			if want := test.want; got != want {
				t.Errorf("Evaluate(%s):\n  got: %v\n want: %v", info, got, want)
			}
		})
	}
}
