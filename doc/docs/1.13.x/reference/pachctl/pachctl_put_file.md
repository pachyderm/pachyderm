## pachctl put file

Put a file into the filesystem.

### Synopsis

Put a file into the filesystem.  This command supports a number of ways to insert data into PFS.

```
pachctl put file <repo>@<branch-or-commit>[:<path/to/file>] [flags]
```

### Examples

```

# Put data from stdin as repo/branch/path:
echo "data" | pachctl put file repo@branch:/path

# Put data from stdin as repo/branch/path and start / finish a new commit on the branch.
echo "data" | pachctl put file -c repo@branch:/path

# Put a file from the local filesystem as repo/branch/path:
pachctl put file repo@branch:/path -f file

# Put a file from the local filesystem as repo/branch/file:
pachctl put file repo@branch -f file

# Put the contents of a directory as repo/branch/path/dir/file:
pachctl put file -r repo@branch:/path -f dir

# Put the contents of a directory as repo/branch/dir/file:
pachctl put file -r repo@branch -f dir

# Put the contents of a directory as repo/branch/file, i.e. put files at the top level:
pachctl put file -r repo@branch:/ -f dir

# Put the data from a URL as repo/branch/path:
pachctl put file repo@branch:/path -f http://host/path

# Put the data from a URL as repo/branch/path:
pachctl put file repo@branch -f http://host/path

# Put the data from an S3 bucket as repo/branch/s3_object:
pachctl put file repo@branch -r -f s3://my_bucket

# Put several files or URLs that are listed in file.
# Files and URLs should be newline delimited.
pachctl put file repo@branch -i file

# Put several files or URLs that are listed at URL.
# NOTE this URL can reference local files, so it could cause you to put sensitive
# files into your Pachyderm cluster.
pachctl put file repo@branch -i http://host/path
```

### Options

```
  -c, --commit                    DEPRECATED: Put file(s) in a new commit.
      --compress                  Compress data during upload. This parameter might help you upload your uncompressed data, such as CSV files, to Pachyderm faster. Use 'compress' with caution, because if your data is already compressed, this parameter might slow down the upload speed instead of increasing.
  -f, --file strings              The file to be put, it can be a local file or a URL. (default [-])
      --header-records uint       the number of records that will be converted to a PFS 'header', and prepended to future retrievals of any subset of data from PFS; needs to be used with --split=(json|line|csv)
  -h, --help                      help for file
  -i, --input-file string         Read filepaths or URLs from a file.  If - is used, paths are read from the standard input.
  -o, --overwrite                 Overwrite the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.
  -p, --parallelism int           The maximum number of files that can be uploaded in parallel. (default 10)
      --progress                  Print progress bars. (default true)
  -r, --recursive                 Recursively put the files in a directory.
      --split line                Split the input file into smaller files, subject to the constraints of --target-file-datums and --target-file-bytes. Permissible values are line, `json`, `sql` and `csv`.
      --target-file-bytes uint    The target upper bound of the number of bytes that each file contains; needs to be used with --split.
      --target-file-datums uint   The upper bound of the number of datums that each file contains, the last file will contain fewer if the datums don't divide evenly; needs to be used with --split.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

