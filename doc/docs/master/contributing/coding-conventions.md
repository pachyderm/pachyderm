# Coding Conventions

All code in this repo should be written in Go, Shell or Make.  Exceptions are
made for things under `examples` because we want to be able to give people
examples of using the product in other languages. And for things like
`doc/conf.py` which configures an outside tool we want to use for docs. However
in choosing outside tooling we prefer tools that we can interface with entirely
using Go. Go's new enough that it's not always possible to find such a tool so
we expect to make compromises on this. In general you should operate under the
assumption that code written in Go, Shell or Make is accessible to all
developers of the project and code written in other languages is accessible to
only a subset and thus represents a higher liability.

## Shell

- https://google.github.io/styleguide/shell.xml
- Scripts should work on macOS as well as Linux.

## Go

Go has pretty unified conventions for style, we vastly prefer embracing these
standards to developing our own.

### Stylistic Conventions

- We have several Go checks that run as part of CI, those should pass. You can
run them with `make pretest` and `make lint`.
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- Command-line flags should use dashes, not underscores.
- Naming
  - Please consider package name when selecting an interface name, and avoid redundancy.
  - e.g.: `storage.Interface` is better than `storage.StorageInterface`.
  - Do not use uppercase characters, underscores, or dashes in package names.
  - Unless there's a good reason, the `package foo` line should match the name
of the directory in which the .go file exists.
  - Importers can use a different name if they need to disambiguate.
  - Locks should be called `lock` and should never be embedded (always `lock
sync.Mutex`). When multiple locks are present, give each lock a distinct name
following Go conventions - `stateLock`, `mapLock` etc.

### Testing Conventions

- All new packages and most new significant functionality must come with test coverage
- Avoid waiting for asynchronous things to happen (e.g. waiting 10 seconds and
assuming that a service will be afterward). Instead you try, wait, retry, etc.
with a limited number of tries. If possible use a method of waiting directly
(e.g. 'flush commit' is much better than repeatedly trying to read from a
commit).

### Go Modules/Third-Party Code

- Go dependencies are managed with go modules (as of 07/11/2019).
- To add a new package or update a package. Do:
  - `go get foo`
    or for a more specific version
    `go get foo@v1.2.3`, `go get foo@master`, `go get foo@e3702bed2`
  - import foo package to you go code as needed.
- Note: Go modules requires you clone the repo outside of the `$GOPATH` or you must pass the `GO111MODULE=on` flag to any go commands. See wiki page on [activating module support](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support)

- See
[The official go modules wiki](https://github.com/golang/go/wiki/Modules)
for more info.

### Docs

- PRs for code must include documentation updates that reflect the changes
that the code introduces.

- When writing documentation, follow the [Style Guide](docs-style-guide.md)
conventions.

- PRs that have only documentation changes, such as typos, is a great place
to start and we welcome your help!

- For most documentation  PRs, you need to `make assets` and push the new
`assets.go` file as well.
