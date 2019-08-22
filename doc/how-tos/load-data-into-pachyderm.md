# Load Your Data Into Pachyderm

Before you read this section, make sure that you are familiar with
the [Data Concepts](../concepts/data-concepts/index.html) and
[Pipeline Concepts](../concepts/pipeline-concepts/index.html).

The data that you commit to Pachyderm is stored in an object store of your
choice, such as Amazon S3, MinIO, Google Cloud Storage, or other. Pachyderm
records the cryptographic hash (`SHA`) of each portion of your data and stores
it as a commit with a unique identifier (ID). This is called content-addressing.
Pachyderm's version control is built on content-addressing technology.
In an object store, data is stored as an unstructured blob that is not
human-readable.
However, Pachyderm enables you to interact with versioned data as you
typically do in a standard disk file system.

Pachyderm stores versioned data in repositories which can contain one or
multiple files, as well as files arranged in directories. Regardless of the
repository structure, Pachyderm versions the state of each data repository
as the data changes over time.

Regardless of how you load your data, every time new data is loaded,
Pachyderm creates a new commit in the corresponding Pachyderm repository.
To put data into Pachyderm, a commit must be *started*, or *opened*.
Data can then be put into Pachyderm as part of that open commit and is
available once the commit is *finished* or *closed*.

Pachyderm provides the following options to load data:

* By using the `pachctl put file` command. This option is great for testing,
development, integration with CI/CD, and for users who prefer scripting.
See [Use `put file`](use-put-file.html).

* By creating a [pipeline](../../concepts/pipeline-concepts/pipeline/index.rst).
Pipelines enable you to do much more than
just putting individual files or folders into Pachyderm. In a pipeline,
you can specify your data source, such as a Pachyderm input repository,
in the `input` field and then configure your code to run automatically
your code as required. The pipeline code can be triggered on-demand or
continuously with the following special types of pipelines:

  * **Spout:** A spout enables you to continuously load
  streaming data from a streaming data source, such as a messaging system
  or message queue into Pachyderm.
  See [Spout](../../concepts/pipeline-concepts/pipeline/spout.html).

  * **Cron:** A cron triggers your pipeline periodically based on the
  interval that you configure in your pipeline spec.
  See [Cron](../../concepts/pipeline-concepts/pipeline/cron.html).

* By using a Pachyderm language client. This option is ideal
for Go, Python, or Scala users who want to push data into Pachyderm from
services or applications written in those languages. If you did not find your
favorite language in the list of supported language clients,
Pachyderm uses a protobuf API which supports many other languages.
See [Pachyderm Language Clients](../../reference/clients.html).

If you are using the Pachyderm Enterprise version, you can use these
additional options:

* By using the S3 gateway. This option is great to use with the existing tools
and libraries that interact with object stores.
See [Using the S3 Gateway](../../enterprise/s3gateway.html).

* By using the Pachyderm dashboard. The Pachyderm Enterprise dashboard
provides a convenient way to upload data right from the UI.
<!--TBA link to the PachHub tutorial-->

  **Note:** In the Pachyderm UI, you can only specify an S3 data source.
  Uploading data from your local device is not supported.

## Load Your Data by Using `pachctl`

To load your data into Pachyderm by using `pachctl`, you first need to create
one or more data repositories. Then, you can use the `pachctl put file`
command to put your data into the created repository.

In Pachyderm, you can *start* and *finish* commits. If you just
run `pachctl put file` and no open commit exists, Pachyderm starts a new
commit, adds the data to which you specified the path in your command, and
finishes the commit. This is called an atomic commit.

Alternatively, you can run `pachctl start commit` to start a new commit.
Then, add your data in multiple iterations, each time appending to the
same commit by running the `pachctl put file` command untill you
close the commit by running `pachctl finish commit`.

To load your data into a repository, complete the following steps:

1. Create a Pachyderm repository:

   ```sh
   $ pachctl create repo <repo name>
   ```

1. Select from the following options:

   * To start and finish an atomic commit, run:

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

     1. Work on your changes, and when ready, put more data:

        ```bash
        $ pachctl put file <repo>@<branch>:</path/to/file2> -f <file2>
        ```

     1. Close the commit:

        ```bash
        $ pachctl finish commit <repo>@<branch>
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
  $ pachctl put file <repo>@<branch>:</path/to/file> -f http://url_path
  ```

* Put data from an object store. You can use `s3://`, `gcs://`, or `as://`
in your filepath:

  ```
  $ pachctl put file <repo>@<branch>:</path/to/file> -f s3://object_store_url
  ```

  **Note:** If you are configuring a local cluster to access an S3 bucket,
  you need to first deploy a Kubernetes `Secret` for the selected object
  store.

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

## Loading Your Data Partially

Depending on your use case, you might decide not to import all of your
data into Pachyderm but only store and apply version control to some
of it. For example, if you have a 10 PB dataset, loading the
whole dataset into Pachyderm is a costly operation that takes
a lot of time and resources. To optimize performance and the
use of resources, you might decide to load some of this data into
Pachyderm, leaving the rest of it in its original source.

One possible way of doing this is adding a `.txt` file with a
URL to the specific file or directory in your dataset to a Pachyderm
repository and refer to that text file in your pipeline. Pachyderm
will not keep versions of the source file, but
it will keep track and provenance of the resulting output commits
in its version-control system.
