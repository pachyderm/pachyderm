package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v54/github"
	flag "github.com/spf13/pflag"
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

var (
	startDay  string
	endDay    string
	authors   []string
	prsToScan []int

	missingJira bool
	lastMinute  bool
	printPRs    bool
	status      bool
	comment     bool
	help        bool
)

func init() {
	flag.CommandLine.SortFlags = false
	flag.StringVar(&startDay, "start-day", "", "Only scan PRs created after the start of the given day (must be YYYY-MM-DD).")
	flag.StringVar(&endDay, "end-day", "", "Only scan PRs created before the start of the given day (must be YYYY-MM-DD).")
	flag.StringSliceVar(&authors, "authors", nil, "Only scan PRs by these authors (comma-separated list of case-sensitive GitHub usernames).")
	flag.IntSliceVar(&prsToScan, "pr-numbers", nil, "Only scan the indicated PRs (comma-separated list of PR numbers).")

	flag.BoolVar(&missingJira, "no-jira-ticket", false, "Act on PRs that are missing a reference to a Jira ticket (e.g. '[CORE-1234]'). Only one of --missing-jira and --last-minute may be set.")
	flag.BoolVar(&lastMinute, "last-minute", false, "Act on PRs that were created after a release sync and before the corresponding release (only since 2023-01-01). Only one of --missing-jira and --last-minute may be set.")
	flag.BoolVar(&printPRs, "print", false, "If true, print all matching PRs scanned, grouped by author, with all PRs matching the given condition starred")
	flag.BoolVar(&status, "add-pr-status", false, "Add a commit status to all PRs scanned (a failing status for those matching --no-jira-ticket or --last-minute, and a passing status to the rest)")
	flag.BoolVar(&comment, "comment", false, "Add a comment to any PRs identified by --no-jira-ticket or --last-minute")
	flag.BoolVarP(&help, "help", "h", false, "Print a help message and exit.")
}

func main() {
	// Parse and validate commandline args
	flag.Parse()
	if help || (!printPRs && !comment && !status) {
		flag.PrintDefaults()
		os.Exit(0)
	}

	var check prCheck
	if missingJira && lastMinute {
		fmt.Fprintf(os.Stderr, "Error: only one of --no-jira-ticket and --last-minute can be set")
		os.Exit(1)
	} else if missingJira {
		check = missingJiraTicket
	} else if lastMinute {
		check = createdLastMinute
	}

	if (comment || status) && !(missingJira || lastMinute) {
		fmt.Fprintf(os.Stderr, "Error: if --comment or --status is set, then one of --no-jira-ticket or --last-minute must be set")
		os.Exit(1)
	}
	var action prAction
	if comment && status {
		fmt.Fprintf(os.Stderr, "Error: at most one of --comment or --status may be set")
		os.Exit(1)
	} else if comment {
		action = commentOnPr
	} else if status {
		action = addStatus
	}
	if len(prsToScan) > 0 && (len(authors) > 0 || startDay != "" || endDay != "") {
		fmt.Fprintf(os.Stderr, "Error: if --pr-numbers is set, then --authors, --start, and --end must not be set")
		os.Exit(1)
	}

	// initialize start and end time (from --start and --end)
	var end, start time.Time
	var err error
	if startDay != "" {
		start, err = time.Parse(time.DateOnly, startDay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse --start-day value %q (must match YYYY-MM-DD): %s", startDay, err)
			os.Exit(1)
		}
	}
	if endDay != "" {
		end, err = time.Parse(time.DateOnly, endDay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse --end-day value %q (must match YYYY-MM-DD): %s", endDay, err)
			os.Exit(1)
		}
	}
	// Clamp 'start' and 'end'. This simplifies date math by ensuring 'start' and
	// 'end' are defined, and 'start' is before the earliest conceivable creation
	// time of a PR, and end is after the the latest conceivable creation time of
	// a PR for Pachyderm.
	earliestStart := earliestPR.Add(-24 * time.Hour)
	if start.Before(earliestStart) {
		start = earliestStart
	}
	latestEnd := time.Now().Add(240 * time.Hour)
	if end.IsZero() || end.After(latestEnd) {
		end = latestEnd
	}
	if start.After(end) {
		fmt.Fprintf(os.Stderr, "PR start day (%s) must be before or on PR end day (%s)", startDay, endDay)
		os.Exit(1)
	}
	authorsMap := make(map[string]bool)
	for _, a := range authors {
		authorsMap[a] = true
	}

	// Initialize GitHub client
	fmt.Println("Starting...")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	fmt.Println("Authing with GitHub...")
	client := github.NewClient(tc)

	// Scan and process PRs, and print the results if requested
	var scannedPRs []*github.PullRequest
	now := time.Now()
	if len(prsToScan) > 0 {
		scannedPRs = getPRs(client, prsToScan)
	} else {
		scannedPRs = scanPRs(client, authorsMap, now, start, end)
	}
	matchingPRs := checkPRs(client, scannedPRs, now, check, action)
	if printPRs {
		printScanned(scannedPRs, matchingPRs)
	}
}
