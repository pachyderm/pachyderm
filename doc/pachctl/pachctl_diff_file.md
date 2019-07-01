## pachctl diff file

Return a diff of two file trees.

### Synopsis


Return a diff of two file trees.

```
pachctl diff file <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>]
```

### Examples

```
```sh

# Return the diff of the file "path" of the repo "foo" between the head of the
# "master" branch and its parent.
$ pachctl diff file foo@master:path

# Return the diff between the master branches of repos foo and bar at paths
# path1 and path2, respectively.
$ pachctl diff file foo@master:path1 bar@master:path2
```
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -s, --shallow           Specifies whether or not to diff subdirectories
```

### Options inherited from parent commands

```
  -v, --verbose   Output verbose logs
```

