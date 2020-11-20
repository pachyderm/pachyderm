package server

import (
	"net/http"
)

type RewriteRoundTripper struct {
	RewriteHost string
}

func RewriteClient(rewriteHost string) *http.Client {
	return &http.Client{
		Transport: RewriteRoundTripper{rewriteHost},
	}
}

func (rt RewriteRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Host = rt.RewriteHost
	return http.DefaultTransport.RoundTrip(req)
}
