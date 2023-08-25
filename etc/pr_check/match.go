package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/go-github/v54/github"
)

// prCheck indicates which of the implemented checks 'actOnMatchingPRs' will
// apply to scanned PRs.
type prCheck byte

const (
	missingJiraTicket prCheck = iota
	createdLastMinute
)

func (c prCheck) String() string {
	switch c {
	case missingJiraTicket:
		return "PR contains Jira ticket"
	case createdLastMinute:
		return "PR reviewed before release"
	}
	return "Error: unrecognized pr check"
}

// prAction indicates which of the implemented actions 'actOnMatchingPRs' will
// apply to scanned PRs.
type prAction byte

const (
	commentOnPr prAction = iota
	addStatus
)

// jiraRE is the regex that a PR's title+body+HEAD branch must match to pass the
// 'missingJiraTicket' check.
var jiraRE = regexp.MustCompile(`\[[A-Z]{2,4}-[0-9]{1,5}\]`)

// firstReleaseSync captures the time of the first release sync since we started
// our Tuesday-based sprint schedule: 2023-07-24 at 11:30am pacific. We
// calculate other release syncs from this.
var firstReleaseSync = time.Date(2023, time.July, 24, 11, 30, 00, 0, mustLoadLocation("America/Los_Angeles"))

// freezeDuration is the length of time after a release sync when fresh PRs
// should not be merged
const freezeDuration = 26 * time.Hour // rough guess
const sprintLength = 14 * 24 * time.Hour

// checkMissingJira checks if the PR's title or body contain an apparent Jira
// ticket (anything matching jiraRE)
func checkMissingJira(pr *github.PullRequest) error {
	if jiraRE.MatchString(pr.GetTitle()) {
		return nil
	}
	if jiraRE.MatchString(pr.GetBody()) {
		return nil
	}
	// Check the PR branch name (also a valid way to link a PR to an issue)
	if jiraRE.MatchString(pr.GetHead().GetRef()) {
		return nil
	}
	return fmt.Errorf("This PR doesn't seem to contain a Jira ticket (neither title nor body match %q)", jiraRE.String())
}

// checkLastMinute checks if the PR's creation time AND the current time are
// both in the window between a sprint's release sync and its release.
func checkLastMinute(pr *github.PullRequest) error {
	releaseSyncToPR := pr.GetCreatedAt().Sub(firstReleaseSync) % sprintLength
	prToNow := time.Now().Sub(pr.GetCreatedAt().Time)
	if releaseSyncToPR+prToNow < freezeDuration {
		return errors.New("This PR may have opened between the most recent release sync and the next upcoming release; please consult with Build & Release before merging")
	}
	return nil
}

func makeComment(client *github.Client, pr *github.PullRequest, comment string) error {
	_, _, err := client.Issues.CreateComment(context.Background(),
		"pachyderm", "pachyderm", pr.GetNumber(), &github.IssueComment{
			Body: &comment,
		})
	return err
}

// makeStatus adds a status to the HEAD commit of 'pr'.
// 'state' must be: pending, success, error, or failure.
func makeStatus(client *github.Client, pr *github.PullRequest, state, description string) error {
	_, _, err := client.Repositories.CreateStatus(context.Background(),
		"pachyderm", "pachyderm", pr.GetHead().GetSHA(), &github.RepoStatus{
			State:       &state,
			Description: &description,
		})
	return err
}

func actOnMatchingPRs(client *github.Client, scannedPRs []*github.PullRequest, check prCheck, action prAction) (matchingPRs map[int]bool) {
	matchingPRs = make(map[int]bool)

	for _, pr := range scannedPRs {
		var err error
		switch check {
		case missingJiraTicket:
			err = checkMissingJira(pr)
		case createdLastMinute:
			err = checkLastMinute(pr)
		}
		if err == nil {
			if action == addStatus {
				// TODO(msteffen): instead of always leaving a status, it may make sense
				// to check whether a failing status has been left previously, and only
				// leave a status in that case.
				makeStatus(client, pr, "success", check.String())
			}
			continue
		}
		matchingPRs[pr.GetNumber()] = true
		switch action {
		case commentOnPr:
			makeComment(client, pr, err.Error())
		case addStatus:
			makeStatus(client, pr, "failure", check.String())
		}
	}
	return matchingPRs
}
