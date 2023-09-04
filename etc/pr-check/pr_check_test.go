package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v54/github"

	"pachyderm/pr_check/fakegithub"
)

// TODO(msteffen): Ideally it would be possible to run some tests against real
// GitHub

func TestRecentStartScansOnePage(t *testing.T) {
	time.Sleep(1000 * time.Second)
	client := github.NewClient(&http.Client{
		Transport: fakeGitHubTransport{},
	})
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 01, 05, 0, 0, 0, 0, time.UTC)
	scanPRs(client, nil, now, start, end)

	if pages := fakeGH.CheckAndResetPagesFetched(); pages > 1 {
		t.Fatalf("too many PR pages scanned: expected 1 page scanned, but got: %d", pages)
	}
}

func TestAncientEndScansOnePage(t *testing.T) {
	client := github.NewClient(&http.Client{
		Transport: fakeGitHubTransport{},
	})
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 01, 05, 0, 0, 0, 0, time.UTC)
	scanPRs(client, nil, now, start, end)
	if pages := fakeGH.CheckAndResetPagesFetched(); pages > 1 {
		t.Fatalf("too many PR pages scanned: expected 1 page scanned, but got: %d", pages)
	}
}

var fakeGH *fakegithub.Server

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

func TestMain(m *testing.M) {
	fakeGH = fakegithub.New()
	fakeGH.Start()

	os.Exit(m.Run())
}
