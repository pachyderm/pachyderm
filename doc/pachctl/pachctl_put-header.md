## pachctl put-header

Put a header file into the filesystem.

### Synopsis


Put-header supports a number of ways to insert data into pfs:
```sh

# Put data from stdin as repo/branch/path:
$ echo "data" | pachctl put-header repo branch path-to-directory

# Put a file from the local filesystem as repo/branch/path:
$ pachctl put-header repo branch path-to-directory -f file

# Delete the existing header:
$ echo "" | pachctl put-header repo branch path -f -

```

```
pachctl put-header repo-name branch [path/to/directory/in/pfs]
```

### Options

```
  -f, --file string   The file to be put, it can be a local file or by default will be read from stdin (default "-")
```

### Options inherited from parent commands

```
      --no-metrics   Don't report user metrics for this command
  -v, --verbose      Output verbose logs
```

