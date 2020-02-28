# Supported Operations

The Pachyderm S3 gateway supports the following operations:

* Create buckets: Creates a repo and branch.
* Delete buckets: Deletes a branch or a repo with all branches.
* List buckets: Lists all branches on all repos as S3 buckets.
* Write objects: Atomically overwrites a file on a branch.
* Remove objects: Atomically removes a file on a branch.
* List objects: Lists the files in the HEAD of a branch.
* Get objects: Gets file contents on a branch.

## List Filesystem Objects

If you have configured your S3 client correctly, you should be
able to see the list of filesystem objects in your Pachyderm
repository by running an S3 client `ls` command.
To list filesystem objects, complete the following steps:

1. Verify that your S3 client can access all of your Pachyderm repositories:

   * If you are using MinIO, type:

     ```bash
     mc ls local
     ```

     **System Response:**

     ```bash
     [2019-07-12 15:09:50 PDT]      0B master.train/
     [2019-07-12 14:58:50 PDT]      0B master.pre_process/
     [2019-07-12 14:58:09 PDT]      0B master.split/
     [2019-07-12 14:58:09 PDT]      0B stats.split/
     [2019-07-12 14:36:27 PDT]      0B master.raw_data/
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600 s3 ls
     ```

     **System Response:**

     ```bash
     2019-07-12 15:09:50 master.train
     2019-07-12 14:58:50 master.pre_process
     2019-07-12 14:58:09 master.split
     2019-07-12 14:58:09 stats.split
     2019-07-12 14:36:27 master.raw_data
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd ls
     ```

     **System Response:**

     ```bash
     2019-07-12 15:09 master.train
     2019-07-12 14:58 master.pre_process
     2019-07-12 14:58 master.split
     2019-07-12 14:58 stats.split
     2019-07-12 14:36 master.raw_data
     ```

1. List the contents of a repository:

   * If you are using MinIO, type:

     ```bash
     mc ls local/master.raw_data
     ```

     **System Response:**

     ```bash
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data
     ```

     **System Response:**

     ```bash
     2019-07-26 11:22:23    2685061 github_issues_medium.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd ls s3://master.raw_data/
     ```

     **System Response:**

     ```bash
     2019-07-26 11:22 2685061 s3://master.raw_data/github_issues_medium.csv
     ```
## Create an S3 Bucket

You can create an S3 bucket in Pachyderm by using the AWS CLI or
the MinIO client commands.
The S3 bucket that you create is a branch in a repository
in Pachyderm.

To create an S3 bucket, complete the following steps:

1. Use the `mb <host/branch.repo>` command to create a new
S3 bucket, which is a repository with a branch in Pachyderm.

   * If you are using MinIO, type:

     ```bash
     mc mb local/master.test
     ```

     **System Response:**

     ```bash
     Bucket created successfully `local/master.test`.
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 mb s3://master.test
     ```

     **System Response:**

     ```bash
     make_bucket: master.test
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd mb s3://master.test
     ```

     This command creates the `test` repository with the `master` branch.

1. Verify that the S3 bucket has been successfully created:

   * If you are using MinIO, type:

     ```bash
     mc ls local
     ```

     **System Response:**

     ```bash
     [2019-07-18 13:32:44 PDT]      0B master.test/
     [2019-07-12 15:09:50 PDT]      0B master.train/
     [2019-07-12 14:58:50 PDT]      0B master.pre_process/
     [2019-07-12 14:58:09 PDT]      0B master.split/
     [2019-07-12 14:58:09 PDT]      0B stats.split/
     [2019-07-12 14:36:27 PDT]      0B master.raw_data/
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 ls
     ```

     **System Response:**

     ```bash
     2019-07-26 11:35:28 master.test
     2019-07-12 14:58:50 master.pre_process
     2019-07-12 14:58:09 master.split
     2019-07-12 14:58:09 stats.split
     2019-07-12 14:36:27 master.raw_data
          ```
   * If you are using S3cmd, type:

     ```bash
     s3cmd ls
     ```

     **System Response:**

     ```bash
     2019-07-26 11:35 master.test
     2019-07-12 14:58 master.pre_process
     2019-07-12 14:58 master.split
     2019-07-12 14:58 stats.split
     2019-07-12 14:36 master.raw_data
     ```

   * You can also use the `pachctl list repo` command to view the
   list of repositories:

     ```bash
     pachctl list repo
     ```

     **System Response:**

     ```bash
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

  ```bash
  mc rb local/master.test
  ```

  **System Response:**

  ```bash
  Removed `local/master.test` successfully.
  ```

* If you are using AWS, type:

  ```bash
  aws --endpoint-url http://localhost:30600/ s3 rb s3://master.test
  ```

  **System Response:**

  ```bash
  remove_bucket: master.test
  ```

* If you are using S3cmd, type:

  ```bash
  s3cmd rb s3://master.test
  ```

## Upload and Download File Objects

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

     ```bash
     mc cp test.csv local/master.raw_data/test.csv
     ```

     **System Response:**

     ```bash
     test.csv:                  62 B / 62 B  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  100.00% 206 B/s 0s
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 cp test.csv s3://master.raw_data
     ```

     **System Response:**

     ```bash
     upload: ./test.csv to s3://master.raw_data/test.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd cp test.csv s3://master.raw_data
     ```

   These commands add the `test.csv` file to the `master` branch in
   the `raw_data` repository. `raw_data` is an input repository.

1. Check that the file was added:

   * If you are using MinIO, type:

     ```bash
     mc ls local/master.raw_data
     ```

     **System Response:**

     ```bash
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     [2019-07-19 12:11:37 PDT]     62B test.csv
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data/
     ```

     **System Response:**

     ```bash
     2019-07-19 12:11:37  2685061 github_issues_medium.csv
     2019-07-19 12:11:37       62 test.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd ls s3://master.raw_data/
     ```

     **System Response:**

     ```bash
     2019-07-19 12:11  2685061 github_issues_medium.csv
     2019-07-19 12:11       62 test.csv
     ```

1. Download a file from MinIO to the
current directory by running the following commands:

   * If you are using MinIO, type:

     ```bash
     mc cp local/master.raw_data/github_issues_medium.csv .
     ```

     **System Response:**

     ```bash
     ...hub_issues_medium.csv:  2.56 MiB / 2.56 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 100.00% 1.26 MiB/s 2s
     ```

   * If you are using AWS, type:

     ```
     aws --endpoint-url http://localhost:30600/ s3 cp s3://master.raw_data/test.csv .
     ```

     **System Response:**

     ```bash
     download: s3://master.raw_data/test.csv to ./test.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd cp s3://master.raw_data/test.csv .
     ```
## Remove a File Object

You can delete a file in the `HEAD` of a Pachyderm branch by using the
MinIO command-line interface:

1. List the files in the input repository:

   * If you are using MinIO, type:

     ```bash
     mc ls local/master.raw_data/
     ```

     **System Response:**

     ```bash
     [2019-07-19 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     [2019-07-19 12:11:37 PDT]     62B test.csv
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data
     ```

     **System Response:**

     ```bash
     2019-07-19 12:11:37    2685061 github_issues_medium.csv
     2019-07-19 12:11:37         62 test.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd ls s3://master.raw_data
     ```

     **System Response:**

     ```bash
     2019-07-19 12:11    2685061 github_issues_medium.csv
     2019-07-19 12:11         62 test.csv
     ```

1. Delete a file from a repository. Example:

   * If you are using MinIO, type:

     ```bash
     mc rm local/master.raw_data/test.csv
     ```

     **System Response:**

     ```bash
     Removing `local/master.raw_data/test.csv`.
     ```

   * If you are using AWS, type:

     ```bash
     aws --endpoint-url http://localhost:30600/ s3 rm s3://master.raw_data/test.csv
     ```

     **System Response:**

     ```bash
     delete: s3://master.raw_data/test.csv
     ```

   * If you are using S3cmd, type:

     ```bash
     s3cmd rm s3://master.raw_data/test.csv
     ```
