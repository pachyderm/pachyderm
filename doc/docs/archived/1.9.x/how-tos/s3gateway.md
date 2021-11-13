# Using the S3Gateway

Pachyderm includes an S3 gateway that enables you to interact with PFS storage
through an HTTP application programming interface (API) that imitates the
Amazon S3 Storage API. Therefore, with Pachyderm S3 gateway, you can interact
with Pachyderm through tools and libraries designed to work with object stores.
For example, you can use these tools:

* [MinIO](https://docs.min.io/docs/minio-client-complete-guide)
* [boto3](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html)

When you deploy `pachd`, the S3 gateway starts automatically.

The S3 gateway has some limitations that are outlined below. If you need richer
access, use the PFS gRPC interface instead, or one of the
[client drivers](https://github.com/pachyderm/python-pachyderm).

## Authentication

If auth is enabled on the Pachyderm cluster, credentials must be passed with
each s3 gateway endpoint using AWS' signature v2 or v4 methods. Object store
tools and libraries provide built-in support for these methods, but they do
not work in the browser. When you use authentication, set the access and secret key
to the same value; they are both the Pachyderm auth token used
to issue the relevant PFS calls.

If auth is not enabled on the Pachyderm cluster, no credentials need to be
passed to s3gateway requests.

## Buckets

The S3 gateway presents each branch from every Pachyderm repository as
an S3 bucket.
For example, if you have a `master` branch in the `images` repository,
an S3 tool sees `images@master` as the `master.images` S3 bucket.

## Versioning

Most operations act on the `HEAD` of the given branch. However, if your object
store library or tool supports versioning, you can get objects in non-HEAD
commits by using the commit ID as the version.

## Port Forwarding

If you do not have direct access to the Kubernetes cluster, you can use port
forwarding instead. Simply run `pachctl port-forward`, which will allow you
to access the s3 gateway through `localhost:30600`.

However, the Kubernetes port forwarder incurs substantial overhead and
does not recover well from broken connections. Connecting to the cluster
directly is therefore faster and more reliable.

## Configure the S3 client

Before you can work with the S3 gateway, configure your S3 client
to access Pachyderm. Complete the steps in one of the sections below that
correspond to your S3 client.

### Configure MinIO

If you are not using the MinIO client, skip this section.

To install and configure MinIO, complete the following steps:

1. Install the MinIO client on your platform as
described on the [MinIO download page](https://min.io/download#/macos).

1. Verify that MinIO components are successfully installed by running
the following command:

   ```shell
   $ minio version
   $ mc version
   Version: 2019-07-11T19:31:28Z
   Release-tag: RELEASE.2019-07-11T19-31-28Z
   Commit-id: 31e5ac02bdbdbaf20a87683925041f406307cfb9
   ```

1. Set up the MinIO configuration file to use the `30600` port for your host:

   ```shell
   vi ~/.mc/config.json
   ```

   You should see a configuration similar to the following:

   * For a minikube deployment, verify the
   `local` host configuration:

     ```shell
     "local": {
               "url": "http://localhost:30600",
               "accessKey": "YOUR-PACHYDERM-AUTH-TOKEN",
               "secretKey": "YOUR-PACHYDERM-AUTH-TOKEN",
               "api": "S3v4",
               "lookup": "auto"
            },
     ```

     Set the access key and secret key to your
     Pachyderm authentication token. If authentication is not enabled
     on the cluster, both parameters must be empty strings.

### Configure the AWS CLI

If you are not using the AWS CLI, skip this section.

If you have not done so already, you need to install and
configure the AWS CLI client on your machine. To configure the AWS CLI,
complete the following steps:

1. Install the AWS CLI for your operating system as described
in the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).

1. Verify that the AWS CLI is installed:

   ```shell
   $ aws --version aws-cli/1.16.204 Python/2.7.16 Darwin/17.7.0 botocore/1.12.194
   ```

1. Configure AWS CLI:

   ```shell
   $ aws configure
   AWS Access Key ID: YOUR-PACHYDERM-AUTH-TOKEN
   AWS Secret Access Key: YOUR-PACHYDERM-AUTH-TOKEN
   Default region name:
   Default output format [None]:
   ```

   Both the access key and secret key should be set to your
   Pachyderm authentication token. If authentication is not enabled
   on the cluster, both parameters must be empty strings.

## Supported Operations

The Pachyderm S3 gateway supports the following operations:

* Create buckets: Creates a repo and branch.
* Delete buckets: Deletes a branch or a repo with all branches.
* List buckets: Lists all branches on all repos as S3 buckets.
* Write objects: Atomically overwrites a file on a branch.
* Remove objects: Atomically removes a file on a branch.
* List objects: Lists the files in the HEAD of a branch.
* Get objects: Gets file contents on a branch.

### List Filesystem Objects

If you have configured your S3 client correctly, you should be
able to see the list of filesystem objects in your Pachyderm
repository by running an S3 client `ls` command.
To list filesystem objects, complete the following steps:

1. Verify that your S3 client can access all of your Pachyderm repositories:

   * If you are using MinIO, type:

     ```shell
     $ mc ls local
     [2019-07-12 15:09:50 PDT]      0B master.train/
     [2019-07-12 14:58:50 PDT]      0B master.pre_process/
     [2019-07-12 14:58:09 PDT]      0B master.split/
     [2019-07-12 14:58:09 PDT]      0B stats.split/
     [2019-07-12 14:36:27 PDT]      0B master.raw_data/
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600 s3 ls
     2019-07-12 15:09:50 master.train
     2019-07-12 14:58:50 master.pre_process
     2019-07-12 14:58:09 master.split
     2019-07-12 14:58:09 stats.split
     2019-07-12 14:36:27 master.raw_data
     ```

1. List the contents of a repository:

   * If you are using MinIO, type:

     ```shell
     $ mc ls local/master.raw_data
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data
     2019-07-26 11:22:23    2685061 github_issues_medium.csv
     ```

### Create an S3 Bucket

You can create an S3 bucket in Pachyderm by using the AWS CLI or
the MinIO client commands.
The S3 bucket that you create is a branch in a repository
in Pachyderm.

To create an S3 bucket, complete the following steps:

1. Use the `mb <host/branch.repo>` command to create a new
S3 bucket, which is a repository with a branch in Pachyderm.

   * If you are using MinIO, type:

     ```shell
     $ mc mb local/master.test
     Bucket created successfully `local/master.test`.
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 mb s3://master.test
     make_bucket: master.test
     ```

1. Verify that the S3 bucket has been successfully created:

   * If you are using MinIO, type:

     ```shell
     $ mc ls local
     [2019-07-18 13:32:44 PDT]      0B master.test/
     [2019-07-12 15:09:50 PDT]      0B master.train/
     [2019-07-12 14:58:50 PDT]      0B master.pre_process/
     [2019-07-12 14:58:09 PDT]      0B master.split/
     [2019-07-12 14:58:09 PDT]      0B stats.split/
     [2019-07-12 14:36:27 PDT]      0B master.raw_data/
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 ls
     2019-07-26 11:35:28 master.test
     2019-07-12 14:58:50 master.pre_process
     2019-07-12 14:58:09 master.split
     2019-07-12 14:58:09 stats.split
     2019-07-12 14:36:27 master.raw_data
          ```

   * You can also use the `pachctl list repo` command to view the
   list of repositories:

     ```shell
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

You can delete an S3 bucket in Pachyderm from the AWS CLI or
MinIO client by running the following command:

* If you are using MinIO, type:

  ```shell
  $ mc rb local/master.test
  Removed `local/master.test` successfully.
  ```

* If you are using AWS, type:

  ```shell
  $ aws --endpoint-url http://localhost:30600/ s3 rb s3://master.test
  remove_bucket: master.test
  ```

### Upload and Download File Objects

For input repositories at the top of your DAG, you can both add files
to and download files from the repository.
When you add files, Pachyderm automatically overwrites the previous
version of the file if it already exists.

Uploading new files is not supported for output repositories,
these are the repositories that are the output of a pipeline.
Not all the repositories that you see in the results of the `ls` command are
input repositories that can be written to. Some of them might be read-only
output repos. Check your pipeline specification to verify which
repositories are the input repos.

To add a file to a repository, complete the following steps:

1. Run the `cp` command for your S3 client:

   * If you are using MinIO, type:

     ```shell
     $ mc cp test.csv local/master.raw_data/test.csv
     test.csv:                  62 B / 62 B  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  100.00% 206 B/s 0s
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 cp test.csv s3://master.raw_data
     upload: ./test.csv to s3://master.raw_data/test.csv
     ```

   These commands add the `test.csv` file to the `master` branch in
   the `raw_data` repository. `raw_data` is an input repository.

1. Check that the file was added:

   * If you are using MinIO, type:

     ```shell
     $ mc ls local/master.raw_data
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     [2019-07-19 12:11:37 PDT]     62B test.csv
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data/
     2019-07-19 12:11:37  2685061 github_issues_medium.csv
     2019-07-19 12:11:37       62 test.csv
     ```

1. Download a file from MinIO to the
current directory by running the following commands:

   * If you are using MinIO, type:

     ```shell
     $ mc cp local/master.raw_data/github_issues_medium.csv .
     ...hub_issues_medium.csv:  2.56 MiB / 2.56 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 100.00% 1.26 MiB/s 2s
     ```

   * If you are using AWS, type:

     ```
     $ aws --endpoint-url http://localhost:30600/ s3 cp s3://master.raw_data/test.csv .
     download: s3://master.raw_data/test.csv to ./test.csv
     ```

### Remove a File Object

You can delete a file in the `HEAD` of a Pachyderm branch by using the
MinIO command-line interface:

1. List the files in the input repository:

   * If you are using MinIO, type:

     ```shell
     $ mc ls local/master.raw_data/
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     [2019-07-19 12:11:37 PDT]     62B test.csv
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data
     2019-07-19 12:11:37    2685061 github_issues_medium.csv
     2019-07-19 12:11:37         62 test.csv
     ```

1. Delete a file from a repository. Example:

   * If you are using MinIO, type:

     ```shell
     $ mc rm local/master.raw_data/test.csv
     Removing `local/master.raw_data/test.csv`.
     ```

   * If you are using AWS, type:

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 rm s3://master.raw_data/test.csv
     delete: s3://master.raw_data/test.csv
     ```

## Unsupported operations

Some of the S3 functionalities are not yet supported by Pachyderm.
If you run any of these operations, Pachyderm returns a standard
S3 `NotImplemented` error.

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
