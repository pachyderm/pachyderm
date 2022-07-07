## pachctl put file

Put a file into the filesystem.

### Synopsis

Put a file into the filesystem.  This command supports a number of ways to insert data into PFS.

```
pachctl put file <repo>@<branch-or-commit>[:<path/to/file>] [flags]
```

### Examples

```

# Put data from stdin at repo@branch:/path
echo "data" | pachctl put file repo@branch:/path

# Put a file from the local filesystem at repo@branch:/file
pachctl put file repo@branch -f file

# Put a file from the local filesystem at repo@branch:/path
pachctl put file repo@branch:/path -f file

# Put the contents of a directory at repo@branch:/dir/file
pachctl put file -r repo@branch -f dir

# Put the contents of a directory at repo@branch:/path/file (without /dir)
pachctl put file -r repo@branch:/path -f dir

# Put the data from a URL at repo@branch:/example.png
pachctl put file repo@branch -f http://host/example.png

# Put the data from a URL at repo@branch:/dir/example.png
pachctl put file repo@branch:/dir -f http://host/example.png

# Put the data from an S3 bucket at repo@branch:/s3_object
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
  -a, --append              Append to the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.
      --compress            Compress data during upload. This parameter might help you upload your uncompressed data, such as CSV files, to Pachyderm faster. Use 'compress' with caution, because if your data is already compressed, this parameter might slow down the upload speed instead of increasing.
  -f, --file strings        The file to be put, it can be a local file or a URL. (default [-])
      --full-path           If true, use the entire path provided to -f as the target filename in PFS. By default only the base of the path is used.
  -h, --help                help for file
  -i, --input-file string   Read filepaths or URLs from a file.  If - is used, paths are read from the standard input.
  -p, --parallelism int     The maximum number of files that can be uploaded in parallel. (default 10)
      --progress            Print progress bars. (default true)
  -r, --recursive           Recursively put the files in a directory.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

