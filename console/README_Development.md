# Development Guide

1. [Coding Standards](https://github.com/pachyderm/company/blob/master/handbook/frontend.md)

## Writing PRs

Your PRs should be concise and descriptive to the contents of your changeset. If following a Jira ticket, your branch should be named after the ticket like `FRON-100`. Our current process is to wait for at least **1** reviewer before merging the PR. Try to break down larger PRs into digestible chunks, if it can be separated into smaller changes. Try to exclude large procedural changes like moving a directory or applying linter changes from new features, and instead publish those changes as clearly marked, independent PRs.

### Do

- Try to provide a brief description of what changes you're bringing in, and why, if necessary.
- Provide screenshots as an easy glimpse into the contents of the PR if there are significant visual components, such as bringing in new UI elements.
- Include any additional details required to be able to see and run the changeset. E.g. Any preliminary setup steps, necessary configurations, or helpful tips.
- Include any details about changes external to the PR. E.g. A link to changes in CI, an example of a bot in action, or a link to a cloud console.
