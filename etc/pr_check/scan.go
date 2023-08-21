package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-github/v54/github"
)

func mustLoadLocation(loc string) *time.Location {
	l, e := time.LoadLocation(loc)
	if e != nil {
		panic(fmt.Sprintf("failed to load location %q: %v", l, e))
	}
	return l
}

// earliestPR is the date of our currently-oldest open PR
// (https://github.com/pachyderm/pachyderm/pull/3607)
var earliestPR = time.Date(2019, 03, 25, 0, 0, 0, 0, mustLoadLocation("America/Los_Angeles"))

func prsSummaryStr(prs []*github.PullRequest) string {
	switch len(prs) {
	case 0:
		return "()"
	case 1:
		return fmt.Sprintf("(%d)", *prs[0].Number)
	case 2:
		return fmt.Sprintf("(%d, %d)", *prs[0].Number, *prs[1].Number)
	case 3:
		return fmt.Sprintf("(%d, %d, %d)", *prs[0].Number, *prs[1].Number, *prs[2].Number)
	default:
		return fmt.Sprintf("(%d, %d, ....)", *prs[0].Number, *prs[1].Number)
	}
}

// scanPRs uses 'client' to request all PRs in the
// 'github.com/pachyderm/pachyderm' repo created between 'start' and 'end'
//
// TODO
// - [ ] scanPRs is only manually tested right now.
// - [ ] scanPRs relies on main() to initialize 'start' and 'end' to sensible
//       values ('start' is the zero time by default, 'end' is some amount of
//       time in the future). This makes 'scanPRs' annoying to test, as the
//       tests would all have to do the same initialization. It should own
//       initialization (main can do basic validation like end < start or
//       something)
// - [ ] Two pieces of logic in here are worth testing, really:
//       - It requests "asc" or "desc" sort order depending on whether 'start'
//         and 'end are "old" (casually defined as "close to the oldest open PRs
//         I've seen). This logic should be factored out and tested
//       - It reads pages until no more are available (GitHub is weird about
//         this--it sets 'next_page' and 'last_page' to zero on the last page.
//         Don't know if this is an API guarantee or what) OR until it's found
//         all PRs in [start, end] (before start with "desc" sort, or after end
//         with "asc" sort).
func scanPRs(client *github.Client, authors map[string]bool, start time.Time, end time.Time) []*github.PullRequest {
	var (
		// direction is the order in which GitHub should return PRs. If the interval
		// [start, end] is close to now, this will be "desc" (most recent PRs
		// first). If the interval [start, end] is close to 'earliestPR', this will
		// be "asc" (oldest PRs first).
		direction string

		// startGap is, very roughly, the amount of time before 'start' and the
		// amount of time after 'end'. Either or both durations may be negative--I
		// made up 'earliestPR', and 'end' is defined to be tomorrow by default..
		startGap, endGap time.Duration

		// Which page of GitHub results is being read
		page = 1

		result []*github.PullRequest
	)

	// Initialize gaps to figure out if [start, end] is "old". Both gaps may be
	// negative (start before 'earliestPR', end in future. main() initializes
	// 'end' to be in the future, so 'endGap' is actually usually negative), but
	// the logic should still be correct.
	startGap = start.Sub(earliestPR)
	endGap = time.Now().Sub(end)

	// Initialize 'direction' and 'stopFn'
	direction = "desc"
	if startGap < endGap {
		direction = "asc"
	}

	// Read data, page by page
forEachPage:
	for {
		fmt.Printf("Getting PRs, page %d...", page)
		nextPrs, resp, err := client.PullRequests.List(context.Background(),
			"pachyderm", "pachyderm",
			&github.PullRequestListOptions{
				Direction: direction,
				ListOptions: github.ListOptions{
					Page: page,
				},
			})
		if err != nil {
			panic("could not request open Pachyderm PRs from GitHub: " + err.Error())
		}
		fmt.Printf("got %d PRs %s\n", len(nextPrs), prsSummaryStr(nextPrs))
		for _, pr := range nextPrs {
			beforeRange := (direction == "asc" && pr.CreatedAt.GetTime().Before(start)) ||
				(direction == "desc" && pr.CreatedAt.GetTime().After(end))
			if beforeRange {
				continue // skip PRs that are before the beginning of [start, end]
			}
			afterRange := (direction == "asc" && pr.CreatedAt.GetTime().After(end)) ||
				(direction == "desc" && pr.CreatedAt.GetTime().Before(start))
			if afterRange {
				break forEachPage
			}
			if len(authors) > 0 && !authors[pr.GetUser().GetLogin()] {
				continue
			}
			result = append(result, pr)
		}
		if page == resp.NextPage || page >= resp.LastPage {
			// no more pages (GH sets 'resp.LastPage' to 0 on the last page, but I
			// don't know if this is guaranteed. nextPage >= resp.LastPage seemed
			// safer)
			break
		}
		page = resp.NextPage
	}
	return result
}

// getPRs just retrieves the specific, given PRs from GitHub, instead of listing
// all PRs and retrieving the ones that match given criteria. Called from main()
// if '--pr-numbers' is set.
func getPRs(client *github.Client, prs []int) []*github.PullRequest {
	var result []*github.PullRequest
	for _, n := range prs {
		pr, _, err := client.PullRequests.Get(context.Background(), "pachyderm", "pachyderm", n)
		if err != nil {
			panic("could not request Pachyderm PR from GitHub: " + err.Error())
		}
		if pr.GetState() == "closed" {
			fmt.Printf("PR #%d is closed..skipping\n", n)
			continue
		}
		fmt.Printf("Got PR #%d\n", n)
		result = append(result, pr)
	}
	return result
}

// printScanned is only called by main() (when the user has passed --print), and
// prints all scanned PRs
func printScanned(scanned []*github.PullRequest, matching map[int]bool) {
	fmt.Println("PRs:")
	userToPRs := make(map[string][]int)
	names := make([]string, 0, 100)
	for _, pr := range scanned {
		author := *pr.User.Login
		allAuthorPRs, ok := userToPRs[author]
		if !ok {
			names = append(names, author)
		}
		userToPRs[author] = append(allAuthorPRs, *pr.Number)
	}
	sort.Strings(names)
	for _, u := range names {
		fmt.Printf("%s: [", u)
		for i, n := range userToPRs[u] {
			if i > 0 {
				fmt.Printf(", ")
			}
			if matching[n] {
				fmt.Printf("*")
			}
			fmt.Printf("%d", n)
		}
		fmt.Println("]")
	}
}
