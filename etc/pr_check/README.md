# A CLI tool for checking PRs

## What this tool does
- Scan all open Pachyderm PRs
- Find any PRs that have one of the specified issues (currently: missing a Jira ticket, or opened too close to the next upcoming release)
- Process the problematic PRs in zero or more of the following ways:
  - Add a failing status check
  - Post a warning comment
  - (TODO) Send the author a Slack message

It's designed to be run standalone, or as a cron job/recurring GitHub action

## Alternatives Considered
- Why not use [github.com/usehaystack/jira-pr-link-action](https://github.com/usehaystack/jira-pr-link-action)?
- Two reasons:
  - I (Matthew) didn't know it existing when I started writing this.
  - I wrote this mainly to address PRs opened too close to the next upcoming release. Detecting missing Jira tickets is marginally useful to me.
    - Incidentally, I'm also not actually sure the above GH action does the right thing--that action is designed to be triggered by `pull_request` events, but a user might add a Jira ticket to their PR's body without pushing any commits. Does that above action re-run, and change the check to passing? I think it would be easy enough for PR authors to rebase and force a re-run, though, so if detecting missing Jira links turns out to be the more useful feature of this script, I would not oppose switching to the above. I just wonder which architecture is better.

## Code Layout
- `main.go`
    - parses commandline arguments, and calls the helper functions in `match.go` and `scan.go`
- `scan.go`
    - scans GitHub for open PRs
    - `scan.go` attempts to request pages in the most efficient order given the values of `--start` and `--end` (if they're old, it'll request PRs sorted oldest-to-newest. If they're recent, it'll do the reverse. It stops requesting PRs once all PRs in its window have been returned)
    - if specific PR numbers are passed, they are read directly
- `match.go`
    - detects which PRs have any of the requested issues, and either comments on them or adds a status check, as requested


