# Using the S3 Gateway

Pachyderm Enterprise includes an S3 gateway that enables you to interact
with PFS storage through an HTTP application programming interface (API)
that imitates the Amazon S3 Storage API. Therefore, with Pachyderm S3
gateway, you can enable the tools and libraries that are designed
to work with object stores, such as MinIO™ and Boto3, to interact with
Pachyderm.

When you deploy `pachd`, the S3 gateway starts automatically. However, the
S3 gateway is an enterprise feature that is only available to paid customers or
during the free trial evaluation. You can access the S3 gateway by pointing
your browser to the following URL:

```bash
http://localhost:30600/
```

Through the S3 gateway, you can only interact with the `HEAD` commits of
your Pachyderm branches that do not require authorization. If you need
to have a more granular access to branches and commits, use the PFS
gRPC Remote Procedure Call interface instead.

## Viewing the List of S3 Buckets

The S3 gateway presents each branch from every Pachyderm repository as an S3 bucket.
For example, if you have a `master` branch in the `images` repository, an S3 tool
sees `images@master` as the `master.images` S3 bucket.

To view the list of S3 buckets, run the following command:

1. Point your browser to `http://<cluster-ip>:30600/`. A list of
S3 buckets appears. Example:

   ![S3 buckets](../_images/s_list_of_buckets.png)

1. Alternatively, you can use port forwarding to connect to the cluster:

   ```
   pachctl port-forward
   ```
   Pachyderm does not recommend to use port forwarding because Kubernetes' port
   forwarder incurs overhead and might not recover well from broken connections.

## Examples of Command-line Operations

The Pachyderm S3 gateway supports the following operations:

* Create buckets: Creates a repo and branch.
* Delete buckets: Deletes a branch or a repo with all branches.
* List buckets: Lists all branches on all repos as S3 buckets.
* Write objects: Atomically overwrites a file on the HEAD of a branch.
* Remove objects: Atomically removes a file on the HEAD of a branch.
* List objects: Lists the files in the HEAD of a branch.
* Get objects: Gets file contents on the HEAD of a branch.

To demonstrate the S3 Gateway functionality, all operations in this section are
performed by using the MinIO command-line client. You can use other
S3 compatible tools to execute the same actions with the corresponding
command-line syntax.

### Configure MinIO

If you have MinIO already installed, skip this section.

To install and configure MinIO, complete the following steps:

1. Install the MinIO client and MinIO server on your platform as
described on the [MinIO download page](https://min.io/download#/macos).
1. Verify that MinIO components are successfully installed by running
the following command:

   ```bash
   $ minio version
   $ mc version
   Version: 2019-07-11T19:31:28Z
   Release-tag: RELEASE.2019-07-11T19-31-28Z
   Commit-id: 31e5ac02bdbdbaf20a87683925041f406307cfb9
   ```

1. Set up the MinIO configuration file to use the `30600` port for the
local host:

   ```bash
   vi ~/.mc/config.json
   ```

   You should see a configuration similar to the following:

   ```bash
   "local": {
             "url": "http://localhost:30600",
             "accessKey": "",
             "secretKey": "",
             "api": "S3v4",
             "lookup": "auto"
         },
   ```

1. Verify that MinIO can access all of your Pachyderm repositories:

   ```bash
   $ mc ls local
   [2019-07-18 13:32:44 PDT]      0B master.test/
   [2019-07-12 15:09:50 PDT]      0B master.train/
   [2019-07-12 14:58:50 PDT]      0B master.pre_process/
   [2019-07-12 14:58:09 PDT]      0B master.split/
   [2019-07-12 14:58:09 PDT]      0B stats.split/
   [2019-07-12 14:36:27 PDT]      0B master.raw_data/
   ```

1. Verify that you can see the contents of a repository:

   ```bash
   $ mc ls local/master.raw_data
   [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
   ```

### Create an S3 Bucket

You can create an S3 bucket in Pachyderm from the MinIO client.
The S3 bucket that you create is a branch in a repository
in Pachyderm.

To create an S3 bucket, complete the following steps:

1. Use the `mc mb <host/branch.repo>` command to create a new
S3 bucket, which is a repository with a branch in Pachyderm.
Example:

   ```bash
   $ mc mb local/master.test
   Bucket created successfully `local/master.test`.
   ```

   This command creates the `test` repository with the `master` branch.

1. Verify that the S3 bucket has been successfully created:

   ```bash
   $ mc ls local
   [2019-07-18 13:32:44 PDT]      0B master.test/
   [2019-07-12 15:09:50 PDT]      0B master.train/
   [2019-07-12 14:58:50 PDT]      0B master.pre_process/
   [2019-07-12 14:58:09 PDT]      0B master.split/
   [2019-07-12 14:58:09 PDT]      0B stats.split/
   [2019-07-12 14:36:27 PDT]      0B master.raw_data/
   ```

   * You can also use the `pachctl list repo` command to view the
   list of repositories:

   ```bash
   $ pachctl list repo
   NAME               CREATED                    SIZE (MASTER)
   test               About an hour ago          0B
   train              6 days ago                 68.57MiB
   pre_process        6 days ago                 1.18MiB
   split              6 days ago                 1.019MiB
   raw_data           6 days ago                 2.561MiB
   ```

   You should see the newly created repository in this list.

### Delete an S3 Bucket

You can delete an S3 bucket in Pachyderm from the MinIO client
by running the following command:

   ```bash
   $ mc rb local/master.test
   Removed `local/master.test` successfully.
   ```

### Upload and Download File Objects

You can add files to and download files from a Pachyderm input repository.
If a file exists in the repository, Pachyderm automatically overwrites it.
This operation is not supported for output repositories. If you try to upload
a file to an output repository, you get an error message:

```
Failed to copy `github_issues_medium.csv`. cannot start a commit on an output
branch
```

Not all the repositories that you see in the output of the `mc ls` command are
input repositories. Check your pipeline specification to verify which
repositories are the PFS input repos.

To add a file to a repository, complete the following steps:

1. Run the `mc cp` command:

   ```
   $ mc cp test.csv local/master.raw_data/test.csv
   test.csv:                  62 B / 62 B  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  100.00% 206 B/s 0s
   ```

   This command adds the `test.csv` file to the `master` branch in
   the `raw_data` repository. `raw_data` is an input repository.

1. Check that the file was added:

   ```bash
   $ mc ls local/master.raw_data
   [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
   [2019-07-19 12:11:37 PDT]     62B test.csv
   ```

1. Download a file to MinIO from a Pachyderm repository by running:

   ```
   $ mc cp local/master.raw_data/github_issues_medium.csv .
   ...hub_issues_medium.csv:  2.56 MiB / 2.56 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 100.00% 1.26 MiB/s 2s
   ```

### Remove a File Object

You can delete a file in the `HEAD` of a Pachyderm branch by using the
MinIO command-line interface by running the following commands:

1. List the files in the iput repository:

     ```bash
    $ mc ls local/master.raw_data/
    [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
    [2019-07-19 12:11:37 PDT]     62B test.csv
    ```
1. Delete a file from a repository. Example:

   <!--- AFAIU, this supposed to work, but it does not.-->

    ```bash
    $ mc rm local/master.raw_data/test.csv
    Removing `local/master.raw_data/test.csv`.
    ```

## Unsupported operations

Some of the S3 functionalities are not yet supported by Pachyderm..
If you run any of these operations, Pachyderm returns a standard
`NotImplemented` error.

The S3 Gateway does not support the following S3 operations:

* Accelerate
* Analytics
* Object copying. PFS supports this functionality through gRPC.
* CORS configuration
* Encryption
* HTML form uploads
* Inventory
* Legal holds
* Lifecycles
* Logging
* Metrics
* Multipart uploads. See writing object documentation above for a workaround.
* Notifications
* Object locks
* Payment requests
* Policies
* Public access blocks
* Regions
* Replication
* Retention policies
* Tagging
* Torrents
* Website configuration

In addition, the Pachyderm S3 gateway has the following limitations:

* No support for authentication or ACLs.
* As per PFS rules, you cannot write to an output repo. At the
moment, Pachyderm returns a 500 error code.
