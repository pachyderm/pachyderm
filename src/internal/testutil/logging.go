package testutil

import (
	"net/http"
	"net/http/httputil"
	"testing"
)

func LogHttpResponse(t testing.TB, response *http.Response, msg string) {
	t.Helper()
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		t.Log("unable to log response body", msg, err, response)
		return
	}
	t.Log(msg, string(dump))
}
