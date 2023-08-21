package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/google/go-github/v54/github"
	"golang.org/x/oauth2"
)

var interestingPRs []int = []int{
	4560, // Support pachctl mount on GKE by wsxiaoys
	8170, // Delete dead code in pach, by seslattery
	6325, // Fix joins example, by me
	7374, // Add tool providing pachctl cli on debug dumps, by PFedak (do we want this?)
	8068, // Add Godoc documentation, by PFedak (pretty sure we want this)
}

var prsToClose []int = []int{
	3607, // Storage layer design, by brycemcanally
}

func jsonStr(v interface{}) string {
	s, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic("could not print Pachyderm PRs from GitHub: " + err.Error())
	}
	return string(s)
}

func prNums(prs []*github.PullRequest) []int {
	var nums []int
	for _, pr := range prs {
		nums = append(nums, *pr.Number)
	}
	return nums
}

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
		return fmt.Sprintf("(%d, %d, %d, ....)", *prs[0].Number, *prs[1].Number, *prs[2].Number)
	}
}

func main() {
	fmt.Println("Starting...")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	fmt.Println("Authing with GitHub...")
	client := github.NewClient(tc)
	prs := make(map[string][]*github.PullRequest)
	for nextPage := 1; ; {
		fmt.Printf("Getting PRs, page %d\n", nextPage)
		nextPrs, resp, err := client.PullRequests.List(context.Background(),
			"pachyderm", "pachyderm",
			&github.PullRequestListOptions{
				ListOptions: github.ListOptions{
					Page: nextPage,
				},
			})
		if err != nil {
			panic("could not request open Pachyderm PRs from GitHub: " + err.Error())
		}
		fmt.Printf("Got %d PRs %s\n", len(nextPrs), prsSummaryStr(nextPrs))
		for _, pr := range nextPrs {
			prs[*pr.User.Login] = append(prs[*pr.User.Login], pr)
		}
		fmt.Printf("cur page: %d, next page: %d, last page: %d\n", nextPage, resp.NextPage, resp.LastPage)
		if nextPage == resp.NextPage || nextPage >= resp.LastPage {
			break
		}
		nextPage = resp.NextPage
	}
	fmt.Println("PRs:")
	var names []string
	for u := range prs {
		names = append(names, u)
	}
	sort.Strings(names) // TODO: use slices.Sort when it's not experimental
	for _, u := range names {
		fmt.Printf("%s: %v\n", u, prNums(prs[u]))
	}
}
