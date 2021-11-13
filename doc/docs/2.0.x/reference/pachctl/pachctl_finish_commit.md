## pachctl finish commit

Finish a started commit.

### Synopsis

Finish a started commit. Commit-id must be a writeable commit.

```
pachctl finish commit <repo>@<branch-or-commit> [flags]
```

### Options

```
      --description string   A description of this commit's contents (synonym for --message)
  -f, --force                finish the commit even if it has provenance, which could break jobs; prefer 'stop job'
  -h, --help                 help for commit
  -m, --message string       A description of this commit's contents (overwrites any existing commit description)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

