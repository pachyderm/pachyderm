package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v54/github"
)

// fakeGitHubTransport is a transport that redirects all requests sent to the
// GitHub API by pr_check to a fake github instance (implemented in
// fake_github.go) instead
type fakeGitHubTransport struct{}

func (g fakeGitHubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.Contains(r.URL.String(), "api.github.com") {
		// Send req to mock server
		r.URL.Scheme = "http"
		r.URL.Host = "localhost:8080"
	}
	return http.DefaultTransport.RoundTrip(r)
}

func TestPagination(t *testing.T) {
	client := github.NewClient(&http.Client{
		Transport: fakeGitHubTransport{},
	})
	// TODO (msteffen): what do to about start/end time?
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	scanPRs(client, nil, startTime, endTime)
	t.Fatal("debug")
}

func TestMain(m *testing.M) {
	newFakeGitHub().start()
	os.Exit(m.Run())
}
