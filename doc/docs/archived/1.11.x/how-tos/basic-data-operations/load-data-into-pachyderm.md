# Ingest Data by Using `pachctl`

## `pachctl put file`

!!! Note
    At any time, run `pachctl put file --help` for the complete list of flags available to you.

1. Load your data into Pachyderm by using `pachctl` requires that one or several input repositories have been created. 

    ```shell
    pachctl create repo <repo name>
    ```

1. Use the `pachctl put file` command to put your data into the created repository. Select from the following options:
    - Atomic commit: no open commit exists in your input repo. Pachyderm automatically starts a new commit, adds your data, and finishes the commit.
    ```shell
    pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
    ```

    - Alternatively, you can manually start a new commit, add your data in multiple `put file` calls, and close the commit by running `pachctl finish commit`.

        1. Start a commit:
            ```shell
            pachctl start commit <repo>@<branch>
            ```
        1. Put your data:
            ```shell
            pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
            ```
        1. Put more data:
            ```shell
            pachctl put file <repo>@<branch>:</path/to/file2> -f <file2>
            ```
        1. Close the commit:
            ```shell
            pachctl finish commit <repo>@<branch>
            ```

## Filepath Formats

In Pachyderm, you specify the path to file by using the `-f` option. A path
to file can be a **local path or a URL to an external resource**. You can add
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
    If you are configuring a local cluster to access an external bucket,
    make sure that Pachyderm has been given the proper access [by configuring your storage credentials](../ingressing_from_diff_cloud)

* Add multiple files at once by using the `-i` option or multiple `-f` flags.
In the case of `-i`, the target file must be a list of files, paths, or URLs
that you want to input all at once:

  ```shell
  pachctl put file <repo>@<branch> -i <file containing list of files, paths, or URLs>
  ```

* Add an entire directory or all of the contents at a particular URL, either
HTTP(S) or object store URL, `s3://`, `gcs://`, and `as://`, by using the
recursive flag, `-r`:

  ```shell
  pachctl put file <repo>@<branch> -r -f <dir>
  ```

## Loading Your Data Partially

Depending on your use case and the volume of your data, 
you might decide to keep your dataset in its original source
and process only a subset in Pachyderm.

Add a metadata file containing a list of URL/path
to your external data to your repo.

Your pipeline code will retrieve the data following their path
without the need to preload it all. 
In this case, Pachyderm will not keep versions of the source file, but it will keep
track and provenance of the resulting output commits. 