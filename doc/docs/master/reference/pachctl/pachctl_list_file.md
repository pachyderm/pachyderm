## pachctl list file

Return the files in a directory.

### Synopsis

Return the files in a directory.

```
pachctl list file <repo>@<branch-or-commit>[:<path/in/pfs>] [flags]
```

### Examples

```

# list top-level files on branch "master" in repo "foo"
pachctl list file foo@master

# list files under directory "dir" on branch "master" in repo "foo"
pachctl list file foo@master:dir

# list top-level files in the parent commit of the current head of "master"
# in repo "foo"
pachctl list file foo@master^

# list top-level files in the grandparent of the current head of "master"
# in repo "foo"
pachctl list file foo@master^2

# list file under directory "dir[1]" on branch "master" in repo "foo"
# : quote and protect regex characters
pachctl list file 'foo@master:dir\[1\]'
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for file
  -o, --output string     Output format when --raw is set: "json" or "yaml" (default "json")
      --raw               Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

