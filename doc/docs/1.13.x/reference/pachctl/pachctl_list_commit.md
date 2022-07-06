## pachctl list commit

Return all commits on a repo.

### Synopsis

Return all commits on a repo.

```
pachctl list commit <repo>[@<branch>] [flags]
```

### Examples

```

# return commits in repo "foo"
pachctl list commit foo

# return commits in repo "foo" on branch "master"
pachctl list commit foo@master

# return the last 20 commits in repo "foo" on branch "master"
pachctl list commit foo@master -n 20

# return commits in repo "foo" since commit XXX
pachctl list commit foo@master --from XXX
```

### Options

```
  -f, --from string       list all commits since this commit
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for commit
  -n, --number int        list only this many commits; if set to zero, list all commits
      --raw               disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

