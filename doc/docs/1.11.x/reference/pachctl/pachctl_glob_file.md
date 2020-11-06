## pachctl glob file

Return files that match a glob pattern in a commit.

### Synopsis

Return files that match a glob pattern in a commit (that is, match a glob pattern in a repo at the state represented by a commit). Glob patterns are documented [here](https://golang.org/pkg/path/filepath/#Match).

```
pachctl glob file <repo>@<branch-or-commit>:<pattern> [flags]
```

### Examples

```

# Return files in repo "foo" on branch "master" that start
# with the character "A".  Note how the double quotation marks around the
# parameter are necessary because otherwise your shell might interpret the "*".
$ pachctl glob file "foo@master:A*"

# Return files in repo "foo" on branch "master" under directory "data".
$ pachctl glob file "foo@master:data/*"
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for file
      --raw               disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

