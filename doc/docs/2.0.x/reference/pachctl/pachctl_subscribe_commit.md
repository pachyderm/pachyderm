## pachctl subscribe commit

Print commits as they are created (finished).

### Synopsis

Print commits as they are created in the specified repo and branch.  By default, all existing commits on the specified branch are returned first.  A commit is only considered 'created' when it's been finished.

```
pachctl subscribe commit <repo>[@<branch>] [flags]
```

### Examples

```

# subscribe to commits in repo "test" on branch "master"
pachctl subscribe commit test@master

# subscribe to commits in repo "test" on branch "master", but only since commit XXX.
pachctl subscribe commit test@master --from XXX

# subscribe to commits in repo "test" on branch "master", but only for new commits created from now on.
pachctl subscribe commit test@master --new
```

### Options

```
      --all               return all types of commits, including aliases
      --from string       subscribe to all commits since this commit
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for commit
      --new               subscribe to only new commits created from now on
      --origin string     only return commits of a specific type
  -o, --output string     Output format when --raw is set: "json" or "yaml" (default "json")
      --raw               Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

