# Configure your S3 client

Before you can work with the S3 gateway, you need to configure your S3 client
to access Pachyderm repo. Complete the steps in one of the sections below that
correspond to your S3 client.

## Configure MinIO
1. Install the MinIO client as
described on the [MinIO download page](https://min.io/download#/macos).

1. Verify that MinIO components are successfully installed by running
the following command:

   ```shell
   minio version
   mc version
   ```

   **System Response:**

   ```
   Version: 2019-07-11T19:31:28Z
   Release-tag: RELEASE.2019-07-11T19-31-28Z
   Commit-id: 31e5ac02bdbdbaf20a87683925041f406307cfb9
   ```

1. Set up the MinIO configuration file to use the S3 Gateway port `30600` for your host:

   ```shell
   vi ~/.mc/config.json
   ```

   You should see a configuration similar to the following:

   * For a minikube deployment, verify the
   `local` configuration:

     ```
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
     on the cluster, you can put any values.

!!! Info
   Find **MinIO** full documentation [here](https://docs.min.io/docs/minio-client-complete-guide).

## Configure the AWS CLI
1. Install the AWS CLI as described
in the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).

1. Verify that the AWS CLI is installed:

   ```shell
   aws --version
   ```

1. Configure AWS CLI. Use the `aws configure` command to configure your credentials file:
   ```shell
   aws configure
   ```
   Both the access key and secret key should be set to your
   Pachyderm authentication token. If authentication is not enabled
   on the cluster, you can pass any value. 

   **System Response:**

   ```
   AWS Access Key ID: YOUR-PACHYDERM-AUTH-TOKEN
   AWS Secret Access Key: YOUR-PACHYDERM-AUTH-TOKEN
   Default region name:
   Default output format [None]:
   ```

!!! Info
   Find **AWS S3 CLI** full documentation [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-services-s3-commands.html).
 
## Configure boto3
Before using Boto3, you need to [set up authentication credentials for your AWS account](#configure-the-AWS-CLI) using the AWS CLI as mentionned previously.

Then follow the [Using boto](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/quickstart.html#using-boto3) documentation starting with importing boto3 in your python file and creating your S3 ressource.!!! Info
   
!!! Info   
   Find **boto3** full documentation [here](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html).