# Coding Conventions

## Languages 

The Pachyderm repository is written using **Go**, **Shell**, and **Make**. Exceptions to this are:

- `/examples`: For showcasing how to use the product in various languages.
- `/doc`: For building documentation using a python-based static site generator ([MkDocs](https://www.mkdocs.org/)).


### Shell

See the [Shell Style Guide](https://google.github.io/styleguide/shellguide.html) for standard conventions. 

### Go

See the [Effective Go Style Guide](https://go.dev/doc/effective_go) for standard conventions.

#### Naming 

- Consider the package name when naming an interface to avoid redundancy. For example, `storage.Interface` is better than `storage.StorageInterface`.
- Do not use uppercase characters, underscores, or dashes in package names.
- The `package foo` line should match the name of the directory in which the `.go` file exists.
- Importers can use a different name if they need to disambiguate.
- Locks should be called `lock` and should never be embedded (always `locksync.Mutex`).
- When multiple locks are present, give each lock a distinct name following Go conventions (e.g., `stateLock`, `mapLock`).


#### Go Modules/Third-Party Code

- See the [Go Modules Usage and Troubleshooting Guide](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support) for managing Go modules.
- Go dependencies are managed with go modules.
- Use `go get foo` to add or update a package; for more specific versions, use  `go get foo@v1.2.3`, `go get foo@master`, or `go get foo@e3702bed2`.

---

## Review

See the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) guide for a list of common comments made on reviews for Go. 


### Checks 

- Run checks using `make lint`. 

### Testing 

- All packages and significant functionality must come with test coverage.
- Avoid waiting for asynchronous things to happen (e.g. waiting 10 seconds and assuming that a service will be afterward). Instead you try, wait, retry, etc. with a limited number of tries. If possible use a method of waiting directly (e.g. 'flush commit' is much better than repeatedly trying to read from a commit).

---

## Documentation

- PRs for code must include documentation updates that reflect the changes that the code introduces.
- When writing documentation, follow the [Style Guide](docs-style-guide.md) conventions.
- PRs that have only documentation changes, such as typos, is a great place to start and we welcome your help!