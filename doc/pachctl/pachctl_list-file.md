## ./pachctl list-file

Return the files in a directory.

### Synopsis


Return the files in a directory.

Examples:

```sh

# list top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master

# list files under directory "dir" on branch "master" in repo "foo"
$ pachctl list-file foo master dir

# list top-level files in the parent commit of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^

# list top-level files in the grandparent of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^2

# list the last n versions of top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master --history n

# list all versions of top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master --history -1

```

```
./pachctl list-file repo-name commit-id path/to/dir
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
      --history int       Return revision history for files.
      --raw               disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

