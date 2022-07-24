## pachctl diff file

Return a diff of two file trees in input repo. Diff of file trees in output repo coming soon.

### Synopsis

Return a diff of two file trees in input repo. Diff of file trees in output repo coming soon.

```
pachctl diff file <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>] [flags]
```

### Examples

```

# Return the diff of the file "path" of the input repo "foo" between the head of the
# "master" branch and its parent.
pachctl diff file foo@master:path

# Return the diff between the master branches of input repos foo and bar at paths
# path1 and path2, respectively.
pachctl diff file foo@master:path1 bar@master:path2
```

### Options

```
      --diff-command string   Use a program other than git to diff files.
      --full-timestamps       Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help                  help for file
      --name-only             Show only the names of changed files.
      --no-pager              Don't pipe output into a pager (i.e. less).
  -s, --shallow               Don't descend into sub directories.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

