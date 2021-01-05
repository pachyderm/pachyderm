## pachctl get file

Return the contents of a file.

### Synopsis

Return the contents of a file.

```
pachctl get file <repo>@<branch-or-commit>:<path/in/pfs> [flags]
```

### Examples

```

# get file "XXX" on branch "master" in repo "foo"
$ pachctl get file foo@master:XXX

# get file "XXX" in the parent of the current head of branch "master"
# in repo "foo"
$ pachctl get file foo@master^:XXX

# get file "XXX" in the grandparent of the current head of branch "master"
# in repo "foo"
$ pachctl get file foo@master^2:XXX
```

### Options

```
  -h, --help              help for file
  -o, --output string     The path where data will be downloaded.
  -p, --parallelism int   The maximum number of files that can be downloaded in parallel (default 10)
  -r, --recursive         Recursively download a directory.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

