package server

import (
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// localhostIdentityServerAddress is the URL we can reach the embedded identity server at
const localhostIdentityServerAddress = "http://localhost:1658/"

// RewriteRoundTripper replaces the expected hostname with a new hostname.
// If a scheme is specified it's also replaced.
type RewriteRoundTripper struct {
	Expected *url.URL
	Rewrite  *url.URL
}

// LocalhostRewriteClient returns an http.Client which replaces the host and scheme
// from `expected` with `localhostIdentityServerAddress`
//
// This helps us work around the case where we are running in hairpin mode.
// (see https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/#a-pod-fails-to-reach-itself-via-the-service-ip)
// We're able to use this to cleverly rewrite requests from pachd -> http://pachd:1658/... as http://localhost:1658/
// so that pachd can talk to the dex server running on the same pod while still making requests to
// 'http://pachd:1658/' which is the configured OIDC Issuer, which preserves OIDC Client-side validation requirements
func LocalhostRewriteClient(expected string) (*http.Client, error) {
	expectedURL, err := url.Parse(expected)
	if err != nil {
		return nil, errors.Errorf("unable to parse URL %q: %w", expected, err)
	}

	rewriteURL, err := url.Parse(localhostIdentityServerAddress)
	if err != nil {
		return nil, errors.Errorf("unable to parse URL %q: %w", localhostIdentityServerAddress, err)
	}

	if rewriteURL.Host == "" || rewriteURL.Scheme == "" {
		return nil, errors.Errorf("invalid URL %q is missing host or scheme", err)
	}

	return &http.Client{
		Transport: RewriteRoundTripper{
			Expected: expectedURL,
			Rewrite:  rewriteURL,
		},
	}, nil
}

// RoundTrip fulfills the http RoundTripper interface
func (rt RewriteRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == rt.Expected.Host {
		req.URL.Host = rt.Rewrite.Host
		req.URL.Scheme = rt.Rewrite.Scheme
	}
	res, err := http.DefaultTransport.RoundTrip(req)
	return res, errors.EnsureStack(err)
}
