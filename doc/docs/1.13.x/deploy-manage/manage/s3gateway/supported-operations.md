# Pachyderm S3 Gateway Supported Operations

The Pachyderm S3 gateway supports the following operations:

* [Create bucket](#create-bucket): Creates a repo and branch.
* [Delete bucket](#delete-bucket): Deletes a branch or a repo with all branches.
* [List buckets](#list-buckets): Lists all branches on all repos as S3 buckets.
* [List objects](#list-objects): Lists the files in the HEAD of a branch.
* [Write object](#write-object) (Upload): Atomically writes a file on a branch of a repo.
* [Get object](#get-object) (Download): Gets file contents on a branch of a repo.
* [Remove object](#remove-object): Atomically removes a file on a branch.

!!! Info

      When using the AWS S3 CLI, simply append `--profile <name-your-profile>` at the end of your command to reference a given profile. If none, the session token will be retrieved from the default profile. More info in the [**Configure your S3 client**](../configure-s3client/#configure-your-s3-client) page. 

      For example, in the **Create Bucket** section below, the command would become:
      ```shell
      aws --endpoint-url http://localhost:30600/ s3 mb s3://master.test --profile <name-your-profile>
      ```

## Create Bucket
Call the *create an S3 bucket* command on your S3 client to create a branch in a Pachyderm repository. 
For example, let's create the `master` branch of the repo `test`.

1. In MinIO, 
    - this would look like:
      ```shell
      $ mc mb local/master.test
      ```
      **System Response:**
      ```
      Bucket created successfully `local/master.test`.
      ```
    - verify that the S3 bucket has been successfully created:
      ```shell
      $ mc ls local
      ```
      **System Response:**
      ```
      [2021-04-26 22:46:08]   0B master.test/
      ```
1. If you are using AWS S3 CLI,
    - this would look like:
      ```shell
      $ aws --endpoint-url http://localhost:30600/ s3 mb s3://master.test
      ```
      **System Response:**
      ```
      make_bucket: master.test
      ```
    - verify that the S3 bucket has been successfully created:
      ```shell
      $ aws --endpoint-url http://localhost:30600/ s3 ls
      ```
      **System Response:**
      ```
      2021-04-26 22:46:08 master.test
      ```
    !!! Note
        Alternatively, You can also use the `pachctl list repo` command to view the
        list of repositories. You should see the newly created repository in this list.
    

## Delete Bucket
Call the *delete an empty S3 bucket* command on your S3 client to delete a Pachyderm repository.

!!! Warning
    The repo must be completely empty.

1. In MinIO, 
    ```shell
    $ mc rb local/master.test
    ```
    **System Response:**
    ```
    Removed `local/master.test` successfully.
    ```
1. If you are using AWS S3 CLI,
    ```shell
    $ aws --endpoint-url http://localhost:30600/ s3 rb s3://master.test
    ```
    **System Response:**
    ```
    remove_bucket: master.test
    ```

## List Buckets
You can check the list of filesystem objects in your Pachyderm
repository by running an S3 client `ls` command.

1. In MinIO,
     ```shell
     $ mc ls local
     ```
     **System Response:**
     ```
     [2021-04-26 15:09:50 PDT]      0B master.train/
     [2021-04-26 14:58:50 PDT]      0B master.pre_process/
     [2021-04-26 14:58:09 PDT]      0B master.split/
     [2021-04-26 14:58:09 PDT]      0B stats.split/
     ```

1. If you are using AWS S3 CLI,

     ```shell
     $ aws --endpoint-url http://localhost:30600 s3 ls
     ```
     **System Response:**
     ```
     2021-04-26 15:09:50 master.train
     2021-04-26 14:58:50 master.pre_process
     2021-04-26 14:58:09 master.split
     2021-04-26 14:58:09 stats.split
     ```

## List Objects
For example, list the contents of the repository raw_data.

1. In MinIO,
     ```shell
     $ mc ls local/master.raw_data
     ```
     **System Response:**
     ```
     [2021-04-26 12:11:37 PDT]  2.6MiB github_issues_medium.csv
     ```

1. If you are using AWS S3 CLI,
     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data
     ```
     **System Response:**
     ```
     2021-04-26  11:22:23    2685061 github_issues_medium.csv
     ```

## Write Object
For example, add the `test.csv` file to the `master` branch in
the `raw_data` repository. `raw_data` being an input repository.

!!! Note
    Not all the repositories that you see in the results of the `ls` command are
    repositories that can be written to. 
    Some of them might be read-only. 
    Note that you should have writting access to the input repo 
    in order to be able to add files to it.

1. In MinIO,
    - this would look like:
        ```shell
        $ mc cp test.csv local/master.raw_data/test.csv
        ```
        **System Response:**
        ```
        test.csv:                  62 B / 62 B  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  100.00% 206 B/s 0s
        ```
    - Check that the file was added:
        ```shell
        $ mc ls local/master.raw_data
        ```
        **System Response:**

        ```
        [2021-04-26 12:11:37 PDT]  2.6MiB github_issues_medium.csv
        [2021-04-26 12:11:37 PDT]     62B test.csv
        ```

1. If you are using AWS S3 CLI,
    - this would look like:
      ```shell
      $ aws --endpoint-url http://localhost:30600/ s3 cp test.csv s3://master.raw_data
      ```
      **System Response:**
      ```
      upload: ./test.csv to s3://master.raw_data/test.csv
      ```
    - Check that the file was added:
      ```shell
      $ aws --endpoint-url http://localhost:30600/ s3 ls s3://master.raw_data/
      ```
      **System Response:**
      ```
      2021-04-26 12:11:37  2685061 github_issues_medium.csv
      2021-04-26 12:11:37       62 test.csv
      ```

## Get Object
For example, download the file `github_issues_medium.csv` from the `master` branch of the repo `raw_data`.

1. In MinIO,
     ```shell
     $ mc cp local/master.raw_data/github_issues_medium.csv .
     ```
     **System Response:**
     ```
     github_issues_medium.csv:  2.56 MiB / 2.56 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 100.00% 1.26 MiB/s 2s
     ```

1. If you are using AWS S3 CLI,
     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 cp s3://master.raw_data/test.csv .
     ```
     **System Response:**
     ```
     download: s3://master.raw_data/test.csv to ./test.csv
     ```

## Remove Object
For example, delete the file `test.csv` in the `HEAD` of the `master` branch of the `raw_data` repo.

1. In MinIO,
     ```shell
     $ mc rm local/master.raw_data/test.csv
     ```
     **System Response:**
     ```
     Removing `local/master.raw_data/test.csv`.
     ```

1. If you are using AWS S3 CLI,

     ```shell
     $ aws --endpoint-url http://localhost:30600/ s3 rm s3://master.raw_data/test.csv
     ```
     **System Response:**
     ```
     delete: s3://master.raw_data/test.csv
     ```
!!! note "See Also:"
    - [Complete S3 Gateway API reference](../../../../reference/s3gateway_api/)