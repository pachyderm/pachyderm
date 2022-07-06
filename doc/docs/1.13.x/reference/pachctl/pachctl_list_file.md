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

# list the last n versions of top-level files on branch "master" in repo "foo"
pachctl list file foo@master --history n

# list all versions of top-level files on branch "master" in repo "foo"
pachctl list file foo@master --history all

# list file under directory "dir[1]" on branch "master" in repo "foo"
# the path is interpreted as a glob pattern: quote and protect regex characters
pachctl list file 'foo@master:dir\[1\]'
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for file
      --history string    Return revision history for files. (default "none")
      --raw               disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

