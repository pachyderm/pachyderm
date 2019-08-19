# Load Your Data by Using `pachctl`

To load your data into Pachyderm by using `pachctl`, you first need to create
one or more data repositories. Then, you can use the `pachctl put file`
command to put your data into the created repository.

Also, in Pachyderm, you can *start* and *finish* commits. If you just
run `pachctl put file` and no open commit exists, Pachyderm starts a new
commit, adds the data to which you specified the path in your command, and
finishes the commit. This is called an atomic commit.

Alternatively, you can run `pachctl start commit` to start a new commit.
Then, add your data in multiple iterations, each time appending to the
same commit by running the `pachctl put file` command untill you
close the commit by running `pachctl finish commit`.

To load your data to a repository, complete the following steps:

1. Create a Pachyderm repository:

   ```sh
   $ pachctl create repo <repo name>
   ```

1. Select from the following options:

   * To start and finish a atomic commit, run:

     ```bash
     $ pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
     ```

   * To start a commit and add data in iterations:

     1. Start a commit:

        ```sh
        $ pachctl start commit <repo>@<branch>
        ```
     1. Put your data:

        ```bash
        $ pachctl put file <repo>@<branch>:</path/to/file1> -f <file1>
        ```

     1. Work on your changes and when ready, put more data:

        ```bash
        $ pachctl put file <repo>@<branch>:</path/to/file2> -f <file2>
        ```

     1. Close the commit:

        ```bash
        $ pachctl finish commit <repo>@<branch>
        ```

## Filepath Format

In Pachyderm, you specify the path to file by using the `-f` option. A path
to file can be a local path or a URL. You can add multiple files or
directories by using the `-i` option. To add the contents of a directory,
use the `-r` flag.

The following table provides examples of `pachctl put file` commands with
various filepaths:

* Put data from a URL:

  ```
  $ pachctl put file <repo>@<branch>:</path/to/file> -f http://url_path
  ```

* Put data from an object store. You can use `s3://`, `gcs://`, or `as://`
in your filepath:

  ```
  $ pachctl put file <repo>@<branch>:</path/to/file> -f s3://object_store_url
  ```

* Add multiple files at once by using the `-i` option or multiple `-f` flags. In
the case of `-i`, the target file must be a list of files, paths, or URLs
that you want to input all at once:

  ```sh
  $ pachctl put file <repo>@<branch> -i <file containing list of files, paths, or URLs>
  ```

* Input data from stdin into a data repository by using a pipe:

  ```sh
  $ echo "data" | pachctl put file <repo>@<branch> -f </path/to/file>
  ```

* Add an entire directory or all of the contents at a particular URL, either
HTTP(S) or object store URL, `s3://`, `gcs://`, and `as://`, by using the
recursive flag, `-r`:

  ```sh
  $ pachctl put file <repo>@<branch> -r -f <dir>
  ```
