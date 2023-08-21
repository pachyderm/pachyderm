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

## Testing
The tests are run against a fake GitHub API, which is implemented in `fake_github.go`. I mostly tried to record responses I got from GitHub for various requests and replay them. Some details:
- The PRs returned are stored in `fake_github_prs.json`. I generated the initial version of that data with:
  ```
  curl -L -v \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/pachyderm/pachyderm/pulls \
    >fake_github_prs.json \
    2> curl_metadata.log
  ```
  - I added PRs to the above to make sure I'd be able to check all cases of the matches in `match.go` that I'm interested in (last-minute PR, non-last-minute PR, PR that was last-minute but now is not, PRs with Jira tickets in each interesting place, etc).
- `curl_metadata.log` includes the headers that the server needs to return. The fake can't be perfectly faithful to github (I don't know how GitHub calculates etags for example; it looks like sha256 + some salt), but the headers that I use are:
  ```
  content-type: application/json; charset=utf-8   # const
  x-github-media-type: github.v3; format=json     # const
  date: Sun, 20 Aug 2023 18:46:10 GMT             # time.Now()
  content-length: 99309                           # len(resp)
  ###
  # The most important header, for pagination:
  # I don't know where 23653453 is coming from, but I just copied it.
  ###
  link: <https://api.github.com/repositories/23653453/pulls?page=2&per_page=5>; rel="next", <https://api.github.com/repositories/23653453/pulls?page=37&per_page=5>; rel="last"
  ```

- time.Now() is used for the time windows (`endGap`, defined in scan.go) and `checkLastMinute` in `match.go`. I'll have to mock time.Now() for the requests to make sense.
