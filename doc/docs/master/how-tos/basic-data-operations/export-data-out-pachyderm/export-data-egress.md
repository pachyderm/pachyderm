# Export Your Data with `egress`

The `egress` field in the Pachyderm [pipeline specification](../../../../reference/pipeline-spec)
enables you to push the results of a pipeline to an
external datastore such as Amazon S3, Google Cloud Storage, or
Azure Blob Storage. After the user code has finished running, but
before the job is marked as successful, Pachyderm pushes the data
to the specified destination.

!!! Note
      Make sure that your cluster has been configured to work with your object store.

Pick the `egress` protocol that applies to your storage:

| Cloud Platform | Protocol | Example |
| -------------- | -------- | ----------- |
| Google Cloud Storage | `gs://` | `gs://gs-bucket/gs-dir` |
| Amazon S3 | `s3://` |  `s3://s3-endpoint/s3-bucket/s3-dir` |
| Azure Blob Storage | `wasb://` | `wasb://default-container@storage-account/az-dir` |

!!! example
    ```json
    "egress": {
       "URL": "s3://bucket/dir"
    },
    ```
