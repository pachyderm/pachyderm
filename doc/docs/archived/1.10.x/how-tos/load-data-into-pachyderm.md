# Load Your Data Into Pachyderm

!!! info
    Before you read this section, make sure that you are familiar with
    the [Data Concepts](../concepts/data-concepts/index.md) and
    [Pipeline Concepts](../concepts/pipeline-concepts/index.md).

The data that you commit to Pachyderm is stored in an object store of your
choice, such as Amazon S3, MinIO, Google Cloud Storage, or other. Pachyderm
records the cryptographic hash (`SHA`) of each portion of your data and stores
it as a commit with a unique identifier (ID). Although the data is
stored as an unstructured blob, Pachyderm enables you to interact
with versioned data as you typically do in a standard file system.

Pachyderm stores versioned data in repositories which can contain one or
multiple files, as well as files arranged in directories. Regardless of the
repository structure, Pachyderm versions the state of each data repository
as the data changes over time.

To put data into Pachyderm, a commit must be *started*, or *opened*.
Data can then be put into Pachyderm as part of that open commit and is
available once the commit is *finished* or *closed*.

Pachyderm provides the following options to load data:

* By using the `pachctl put file` command. This option is great for testing,
development, integration with CI/CD, and for users who prefer scripting.
See [Load Your Data by Using `pachctl`](#load-your-data-by-using-pachctl).

* By creating a special type of pipeline that pulls data from an
outside source.
Because Pachyderm pipelines can be any arbitrary code that runs
in a Docker container, you can call out to external APIs or data
sources and pull in data from there. Your pipeline code can be
triggered on-demand or
continuously with the following special types of pipelines:

  * **Spout:** A spout enables you to continuously load
  streaming data from a streaming data source, such as a messaging system
  or message queue into Pachyderm. 
  See [Spout](../concepts/pipeline-concepts/pipeline/spout.md).

  * **Cron:** A cron triggers your pipeline periodically based on the
  interval that you configure in your pipeline spec.
  See [Cron](../concepts/pipeline-concepts/pipeline/cron.md).

  **Note:** Pipelines enable you to do much more than just ingressing
  data into Pachyderm. Pipelines can run all kinds of data transformations
  on your input data sources, such as a Pachyderm repository, and be
  configured to run your code automatically as new data is committed.
  For more information, see
  [Pipeline](../concepts/pipeline-concepts/pipeline/index.md).

* By using a Pachyderm language client. This option is ideal
for Go or Python users who want to push data into Pachyderm from
services or applications written in those languages. If you did not find your
favorite language in the list of supported language clients,
Pachyderm uses a protobuf API which supports many other languages.
See [Pachyderm Language Clients](../reference/clients.md).

If you are using the Pachyderm Enterprise version, you can use these
additional options:

* By using the S3 gateway. This option is great to use with the existing tools
and libraries that interact with S3-compatible object stores.
See [Using the S3 Gateway](../../deploy-manage/manage/s3gateway/).

* By using the Pachyderm dashboard. The Pachyderm Enterprise dashboard
provides a convenient way to upload data right from the UI.
<!--TBA link to the PachHub tutorial-->

!!! note
    In the Pachyderm UI, you can only specify an S3 data source.
    Uploading data from your local device is not supported.

## Load Your Data by Using `pachctl`

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
