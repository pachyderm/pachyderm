package server

import (
	"fmt"
	"net/http"
	"net/url"
)

// RewriteRoundTripper replaces the expected hostname with a new hostname.
// If a scheme is specified it's also replaced.
type RewriteRoundTripper struct {
	Expected *url.URL
	Rewrite  *url.URL
}

func RewriteClient(expected, rewrite string) (*http.Client, error) {
	expectedUrl, err := url.Parse(expected)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL %q: %w", expected, err)
	}

	rewriteUrl, err := url.Parse(rewrite)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL %q: %w", rewrite, err)
	}
	return &http.Client{
		Transport: RewriteRoundTripper{
			Expected: expectedUrl,
			Rewrite:  rewriteUrl,
		},
	}, nil
}

func (rt RewriteRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == rt.Expected.Host {
		req.URL.Host = rt.Rewrite.Host
		if rt.Rewrite.Scheme != "" {
			req.URL.Scheme = rt.Rewrite.Scheme
		}
	}
	return http.DefaultTransport.RoundTrip(req)
}
