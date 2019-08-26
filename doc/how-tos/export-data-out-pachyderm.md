# Export Your Data From Pachyderm

After you build a pipeline, you most probably want to see the
results that the pipeline has produced. Every commit into an
input repository results in a corresponding commit into an
output repository.

To access the results of
a pipeline, you can use one of the following methods:

* By running the `pachctl get file` command. This
icommand returns the contents of the specified file.<br>
See [Export Your Data with `pachctl`](#export-your-data-with-pachctl).<br>

* By configuring the pipeline. You can configure the following
data exporting methods in a Pachyderm pipeline:

  * An `egress` property in the `output_branch` field that enables you
  to export your data to an external datastore, such as Amazon S3,
  Google Cloud Storage, and others.<br>
  See [Export data by using `egress`](#export-your-data-with-egress).<br>

  * A service. A Pachyderm service exposes the results of the
  pipeline processing on a specific port in the form of a dashboard
  or similar.<br>
  See [Service](../concepts/pipeline-concepts/pipeline/service.html).<br>

* By using the S3 gateway. Pachyderm Enterprise users can reuse
  their existing tools and libraries that work with object store
  to export their data with the S3 gateway.<br>
  See [Using the S3 Gateway](../enterprise/s3gateway.html).

## Export Your Data with `pachctl`

The `pachctl get file` command enables you to get the contents
of a file in a Pachyderm repository. You need to know the file
path to specify it in the command.

To export your data with pachctl:

1. Get the list of files in the repository:

   ```bash
   $ pachctl list file <repo>@<branch-or-commit>
   ```

   **Example:**

   ```bash
   $ pachctl list commit images@d20c41595f804cb090ea2e79e432af15
   REPO   BRANCH COMMIT                           PARENT                           STARTED           DURATION           SIZE
   images master d20c41595f804cb090ea2e79e432af15 91cde2c8ef0a47c5ab19c0251c1f9e59 About an hour ago Less than a second 115.3KiB
   images master 91cde2c8ef0a47c5ab19c0251c1f9e59 265a73951fe94bf4a5a46d4f92f0d17e About an hour ago Less than a second 114.9KiB
   images master 265a73951fe94bf4a5a46d4f92f0d17e 2e24cfca41d0459f85061221146a3ce2 About an hour ago Less than a second 57.64KiB
   images master 2e24cfca41d0459f85061221146a3ce2 <none>                           2 hours ago       Less than a second 57.27KiB
   ```

1. Get the contents of a specific file:

   ```bash
   pachctl get file <repo>@<branch-or-commit>:<path/to/file>
   ```

   **Example:**

   ```bash
   $ pachctl get file images@master:user_data.csv
   1,cyukhtin0@stumbleupon.com,144.155.176.12
   2,csisneros1@over-blog.com,26.119.26.5
   3,jeye2@instagram.com,13.165.230.106
   4,rnollet3@hexun.com,58.52.147.83
   5,bposkitt4@irs.gov,51.247.120.167
   6,vvenmore5@hubpages.com,161.189.245.212
   7,lcoyte6@ask.com,56.13.147.134
   8,atuke7@psu.edu,78.178.247.163
   9,nmorrell8@howstuffworks.com,28.172.10.170
   10,afynn9@google.com.au,166.14.112.65
   ```

   Also, you can view the parent, grandparent, and any previous
   revision by using the caret (`^`) symbol with a number that
   corresponds to an ancestor in sequence:

   * To view a parent of a commit:

     ```bash
     $ pachctl get file <repo>@<branch-or-commit>^:<path/to/file>
     ```

   * To view an `<n>` parent of a commit:

     ```bash
     $ pachctl get file <repo>@<branch-or-commit>^<n>:<path/to/file>
     ```

     **Example:**

     ```bash
     $ pachctl get file images@master^4:user_data.csv
     ```

     You can specify any number in the `^<n>` notation. If the file
     exists in that commit, Pachyderm returns it. If the file
     does not exist in that revision, Pachyderm displays the following
     message:

     ```
     $ pachctl get file <repo>@<branch-or-commit>^<n>:<path/to/file>
     file "<path/to/file>" not found
     ```

## Export Your Data with `egress`

The `egress` field in the Pachyderm [pipeline specification](../reference/pipeline_spec.html)
enables you to push the results of a pipeline to an
external datastore such as Amazon S3, Google Cloud Storage, or
Azure Blob Storage. After the user code has finished running, but
before the job is marked as successful, Pachyderm pushes the data
to the specified destination.

You can specify the following `egress` protocols for the
corresponding storage:

**Note:** Use the horizontal scroll bar in the table below
to view full descriptions and syntax.

| Cloud Platform | Protocol | Description |
| -------------- | -------- | ----------- |
| Google Cloud <br>Storage | `gs://` | GCP uses the utility called `gsutil` to access GCP storage resources <br> from a CLI. This utility uses the `gs://` prefix to access these resources. <br>**Example:**<br> `gs://gs-bucket/gs-dir` |
| Amazon S3 | `s3://` | The Amazon S3 storage protocol requires you to specify an `s3://`<br>prefix before the address of an Amazon resource. A valid address must <br>include an endpoint and a bucket, and, optionally, a directory in your <br>Amazon storage. <br>**Example:**<br> `s3://s3-endpoint/s3-bucket/s3-dir` |
| Azure Blob <br>Storage | `wasb://` | Microsoft Windows Azure Storage Blob (WASB) is the default Azure <br>filesystem that outputs your data through `HDInsight`. To output your <br>data to Azure Blob Storage, use the ``wasb://`` prefix, the container name, <br>and your storage account in the path to your directory. <br>**Example:**<br>`wasb://default-container@storage-account/az-dir` |

**Example:**

```
"output_branch": string,
 "egress": {
    "URL": "s3://bucket/dir"
 },
```
