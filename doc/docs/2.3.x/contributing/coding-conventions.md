# Coding Conventions

Interested in contributing to Pachyderm's code? Learn the conventions here! For setup instructions, see [Setup for Contributors](./setup.md).

## Languages 

The Pachyderm repository is written using **Go**, **Shell**, and **Make**. Exceptions to this are:

- `/examples`: For showcasing how to use the product in various languages.
- `/doc`: For building documentation using a python-based static site generator ([MkDocs](https://www.mkdocs.org/)).


### Shell

- See the [Shell Style Guide](https://google.github.io/styleguide/shellguide.html) for standard conventions. 
- Add [`set -eou pipefail`](https://explainshell.com/explain?cmd=set+-euo+pipefail) to your scripts.

### Go

See the [Effective Go Style Guide](https://go.dev/doc/effective_go) for standard conventions.

#### Naming 

- Consider the package name when naming an interface to avoid redundancy. For example, `storage.Interface` is better than `storage.StorageInterface`.
- Do not use uppercase characters, underscores, or dashes in package names.
- The `package foo` line should match the name of the directory in which the `.go` file exists.
- Importers can use a different name if they need to disambiguate.
- When multiple locks are present, give each lock a distinct name following Go conventions (e.g., `stateLock`, `mapLock`).


#### Go Modules/Third-Party Code

- See the [Go Modules Usage and Troubleshooting Guide](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support) for managing Go modules.
- Go dependencies are managed with go modules.
- Use `go get foo` to add or update a package; for more specific versions, use  `go get foo@v1.2.3`, `go get foo@master`, or `go get foo@e3702bed2`.

### YAML

- See the [Helm Best Practices](https://helm.sh/docs/chart_best_practices/conventions/) guide series. 

---

## Review

- See the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) guide for a list of common comments. 
- See the [Go Test Comments](https://github.com/golang/go/wiki/TestComments) guide for a list of common test code comments.
- Make sure CI is passing for your branch.


### Checks 

- Run checks using `make lint`. 

### Testing 

- All packages and significant functionality must come with test coverage.
- Local unit tests should pass before pushing to GitHub (`make localtest` or `make integration-tests` for integrations).
- Use short flag for local tests only. 
- Avoid waiting for asynchronous things to happen; If possible, use a method of waiting directly (e.g. 'flush commit' is much better than repeatedly trying to read from a commit).
- Run single tests or tests from a single package; the Go tool only supports tests that match a regular expression (for example, `go test -v ./src/path/to/package -run ^TestMyTest`).

---

## Documentation

- When writing documentation, follow the [Style Guide](docs-style-guide.md) conventions.
- PRs that have only documentation changes, such as typos, is a great place to start and we welcome your help!
