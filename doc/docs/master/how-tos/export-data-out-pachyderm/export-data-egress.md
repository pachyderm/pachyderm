# Export Your Data with `egress`

The `egress` field in the Pachyderm [pipeline specification](../../../reference/pipeline_spec)
enables you to push the results of a pipeline to an
external datastore such as Amazon S3, Google Cloud Storage, or
Azure Blob Storage. After the user code has finished running, but
before the job is marked as successful, Pachyderm pushes the data
to the specified destination.

You can specify the following `egress` protocols for the
corresponding storage:

| Cloud Platform | Protocol | Description |
| -------------- | -------- | ----------- |
| Google Cloud <br>Storage | `gs://` | GCP uses the utility called `gsutil` to access GCP storage resources <br> from a CLI. This utility uses the `gs://` prefix to access these resources. <br>**Example:**<br> `gs://gs-bucket/gs-dir` |
| Amazon S3 | `s3://` | The Amazon S3 storage protocol requires you to specify an `s3://`<br>prefix before the address of an Amazon resource. A valid address must <br>include an endpoint and a bucket, and, optionally, a directory in your <br>Amazon storage. <br>**Example:**<br> `s3://s3-endpoint/s3-bucket/s3-dir` |
| Azure Blob <br>Storage | `wasb://` | Microsoft Windows Azure Storage Blob (WASB) is the default Azure <br>filesystem that outputs your data through `HDInsight`. To output your <br>data to Azure Blob Storage, use the ``wasb://`` prefix, the container name, <br>and your storage account in the path to your directory. <br>**Example:**<br>`wasb://default-container@storage-account/az-dir` |

!!! example
    ```json
    "egress": {
       "URL": "s3://bucket/dir"
    },
    ```
