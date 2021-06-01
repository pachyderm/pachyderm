# Load Your Data by Using `pachctl`

## `pachctl put file`

The `pachctl put file` command enables you to do everything from
loading local files into Pachyderm to pulling data from an existing object
store bucket and extracting data from a website. With
`pachctl put file`, you can append new data to the existing data or
overwrite the existing data. All these options can be configured by using
the flags available with this command. Run `pachctl put file --help` to
view the complete list of flags that you can specify.

To load your data into Pachyderm by using `pachctl`, you first need to create
one or more data repositories. Then, you can use the `pachctl put file`
command to put your data into the created repository.

In Pachyderm, you can *start* and *finish* commits. If you just
run `pachctl put file` and no open commit exists, Pachyderm starts a new
commit, adds the data to which you specified the path in your command, and
finishes the commit. This is called an atomic commit.

Alternatively, you can run `pachctl start commit` to start a new commit.
Then, add your data in multiple `put file` calls, and finally, when ready,
close the commit by running `pachctl finish commit`.

To load your data into a repository, complete the following steps:

1. Create a Pachyderm repository:

    ```shell
    pachctl create repo <repo name>
    ```

1. Select from the following options:

   * To start and finish an atomic commit, run:

     ```shell
     pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
     ```

   * To start a commit and add data in iterations:

     1. Start a commit:

        ```shell
        pachctl start commit <repo>@<branch>
        ```
     1. Put your data:

        ```shell
        pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
        ```

     1. Work on your changes, and when ready, put more data:

        ```shell
        pachctl put file <repo>@<branch>:</path/to/file2> -f <file2>
        ```

     1. Close the commit:

        ```shell
        pachctl finish commit <repo>@<branch>
        ```

## Filepath Format

In Pachyderm, you specify the path to file by using the `-f` option. A path
to file can be a local path or a URL to an external resource. You can add
multiple files or directories by using the `-i` option. To add contents
of a directory, use the `-r` flag.

The following table provides examples of `pachctl put file` commands with
various filepaths and data sources:

* Put data from a URL:

  ```
  pachctl put file <repo>@<branch>:</path/to/file> -f http://url_path
  ```

* Put data from an object store. You can use `s3://`, `gcs://`, or `as://`
in your filepath:

  ```
  pachctl put file <repo>@<branch>:</path/to/file> -f s3://object_store_url
  ```

!!! note
    If you are configuring a local cluster to access an S3 bucket,
    you need to first deploy a Kubernetes `Secret` for the selected object
    store.

* Add multiple files at once by using the `-i` option or multiple `-f` flags.
In the case of `-i`, the target file must be a list of files, paths, or URLs
that you want to input all at once:

  ```shell
  pachctl put file <repo>@<branch> -i <file containing list of files, paths, or URLs>
  ```

* Input data from stdin into a data repository by using a pipe:

  ```shell
  echo "data" | pachctl put file <repo>@<branch> -f </path/to/file>
  ```

* Add an entire directory or all of the contents at a particular URL, either
HTTP(S) or object store URL, `s3://`, `gcs://`, and `as://`, by using the
recursive flag, `-r`:

  ```shell
  pachctl put file <repo>@<branch> -r -f <dir>
  ```

## Loading Your Data Partially

Depending on your use case, you might decide not to import all of your
data into Pachyderm but only store and apply version control to some
of it. For example, if you have a 10 PB dataset, loading the
whole dataset into Pachyderm is a costly operation that takes
a lot of time and resources. To optimize performance and the
use of resources, you might decide to load some of this data into
Pachyderm, leaving the rest of it in its original source.

One possible way of doing this is by adding a metadata file with a
URL to the specific file or directory in your dataset to a Pachyderm
repository and refer to that file in your pipeline.
Your pipeline code would read the URL or path in the external data
source and retrieve that data as needed for processing instead of
needing to preload it all into a Pachyderm repo. This method works
particularly well for mostly immutable data because in this case,
Pachyderm will not keep versions of the source file, but it will keep
track and provenance of the resulting output commits in its
version-control system.
