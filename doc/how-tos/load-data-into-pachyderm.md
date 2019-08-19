# Load Your Data Into Pachyderm

Before you read this section, make sure that you are familiar with
the [Data Concepts]() and [Pipeline Concepts]().

The data that you commit to Pachyderm is stored in object storage of your
choice, such as Amazon S3, MinIO, Google Cloud Storage, or other. Pachyderm
records the cryptographic hash (SHA) of each portion of your data and stores
it as a commit with a unique identifier (ID). This is called content-addressing.
Pachyderm's version control is built on content-addressing. In an object store,
data is stored as an unstructured blob that is not human-readable.
However, Pachyderm enables you to interact with versioned data like you
typically do in a normal file system.

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

* By creating a spout. A spout enables you to continuously load
streaming data from a streaming data source, such as a messaging system
or message queue into Pachyderm.

* By using a Pachyderm language client. This option is ideal
for Go, Python, or Scala users who want to push data into Pachyderm from
services or applications written in those languages. If you did not find your
favorite language in the list of supported language clients,
Pachyderm uses a protobuf API which supports many other languages.

If you are using the Pachyderm Enterprise version, you can use these
additional options:

* By using the S3 gateway. This option is great to use with the existing tools
and libraries that interact with object stores.
* By using the Pachyderm dashboard. The Pachyderm Enterprise dashboard
provides a convenient way to upload data right from the UI.
