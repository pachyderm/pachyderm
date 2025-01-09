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
	// Populate fakeGH with realistic-looking PRs:
	// fakePRs, err := os.Open("fake_github_prs.json")
	// if err != nil {
	// 	panic("could not open fake_github_prs.json:\n" + err.Error())
	// }
	// json.NewDecoder(fakePRs).Decode(&r.prs)
	/* >>> */
	// create 6 PRs separated by each of the durations below (so some recent PRs
	// close together, some older PRs farther apart, some PRs that are older still
	// and more sparse, etc. Roughly matches the current distribution in GitHub,
	// but stable for testing
	dts := []time.Duration{
		180 * 24 * time.Hour,
		90 * 24 * time.Hour,
		30 * 24 * time.Hour,
		14 * 24 * time.Hour,
		7 * 24 * time.Hour,
		60 * time.Hour,
		24 * time.Hour,
		6 * time.Hour,
	}
	created := earliestPR
	fakeGH.AddPR("", "", &fakegithub.User{Login: "alice", Type: "User"}, created)
	for _, dt := range dts {
		for i := 0; i < 5; i++ {
			created = created.Add(dt)
			fakeGH.AddPR(
				"Title title title title tiiitle",
				"body body body",
				&fakegithub.User{Login: "bob", Type: "User"},
				created)
		}
	}
	fakeGH.Start()

	os.Exit(m.Run())
}
