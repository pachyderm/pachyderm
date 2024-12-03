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
  - I (Matthew) didn't know it existed when I started writing this.
  - I wrote this mainly to address PRs opened too close to the next upcoming release. Detecting missing Jira tickets is marginally useful to me.
    - Incidentally, I'm also not actually sure the above GH action does the right thing--that action is designed to be triggered by `pull_request` events, but a user might add a Jira ticket to their PR's body without pushing any commits. Does that above action re-run, and change the check to passing? I think it would be easy enough for PR authors to rebase and force a re-run, though, so if detecting missing Jira links turns out to be the more useful feature of this script, I would not oppose switching to the above. I just wonder which architecture is better.

## Code Layout
```
.
├── go.mod     - deps
├── go.sum     - deps
├── README.md  - this file
├── main.go    - entrypoint for the pr-check CLI tool
├── scan.go    - implements scanning github PRs ("asc" vs. "desc", collating pages, etc.)
├── check.go   - implements the various checks (--no-jira-ticket, --last-minute) and actions
│                (--status, --comment)
├── scan_test.go    - tests scan.go (chooses "asc" vs. "desc" correctly; doesn't scan more pages
│                     than needed, etc.)
├── test_data.json  - test data loaded into fake github by scan_test.go
│
└── fakegithub          - A fake (local) implementation of a few GitHub API calls, for testing
    ├── fake_github.go  - Muxer logic, serving, etc
    ├── list.go         - handles '/pulls'
    ├── generate_pr.go  - Loads and executes templates for responding to the above (there's a lot
    │                     of boilerplate in responses)
    └── templates       - Actual template text for the above
        ├── pr.json.tmpl    - template response to `/pulls`
        ├── repo.json.tmpl  - template for 'user'-type fields in pr.json.tmpl
        └── user.json.tmpl  - template for 'repo'-type fields in pr.json.tmpl
```

## Testing
The tests are run against a fake GitHub API, which is implemented in `fakegithub/`. I originally tried to record responses I got from GitHub for various requests and replay them, but I quickly needed to start generating synthetic PR responses to test `match.go`.

Now, I have templates for github responses that take a small number of arguments, and store template values in `test_data.json` or else populate them directly in tests.

Some details:
- `test_data.json` is generated like so:
  ```
  $ curl -L -v \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/pachyderm/pachyderm/pulls \
    >response.json \
    2> metadata.log

  # Process the downloaded PRs into a form that the tests can consume:
  # - Reverse the order of PRs (which is "desc" by default)
  # - Remove IDs and boilerplate (fake_github will generate that)
  # (but still use recent PRs)
  $ jq '.[] | reverse | { "number":.number, "title":.title, "body":.body, "created":.created_at, "author":.user.login }' response.json >test_data.json
  ```
- Note that, at this point, regenerating `test_data.json` will likely break `scan_test.go`. Maybe that will make sense at some point, but the above is mostly intended to be informative.
- `metadata.log` from the command above includes the headers that the server needs to return. The fake can't really be perfectly faithful to github (I don't know how GitHub calculates etags for example; it looks like sha256 + some salt), but the headers that I emulate are:
  ```
  content-type: application/json; charset=utf-8   # const
  x-github-media-type: github.v3; format=json     # const
  date: Sun, 20 Aug 2023 18:46:10 GMT             # time.Now()
  content-length: 99309                           # len(resp)
  ###
  # The most important header, for pagination:
  ###
  # page=1
  link: <https://api.github.com/repositories/23653453/pulls?page=2&per_page=40>; rel="next", <https://api.github.com/repositories/23653453/pulls?page=5&per_page=40>; rel="last"
  # page=2
  link: <https://api.github.com/repositories/23653453/pulls?page=1&per_page=40>; rel="prev", <https://api.github.com/repositories/23653453/pulls?page=3&per_page=40>; rel="next", <https://api.github.com/repositories/23653453/pulls?page=5&per_page=40>; rel="last", <https://api.github.com/repositories/23653453/pulls?page=1&per_page=40>; rel="first"
  # page=5
  link: <https://api.github.com/repositories/23653453/pulls?page=4&per_page=40>; rel="prev", <https://api.github.com/repositories/23653453/pulls?page=1&per_page=40>; rel="first"
  ```

## Future Work
- Send people slack messages as a lighter-weight alternative to comments and status checks
- Once the above has been implemented, automatically notify people about PRs they're blocking (tests need to be rerun, waiting on review, ready to merge, etc.)
- Check whether the Jira ticket linked by a PR has a release version attached to it, and alert if not
  - Eventually: automatically set the release version if not: next minor release for PRs in master, or next patch release for PRs in a release branch
- Automatically cherry-pick PRs made on master into release branches if requested
