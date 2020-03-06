# Configure the S3 client

Before you can work with the S3 gateway, you need to configure your S3 client
to access Pachyderm. Complete the steps in one of the sections below that
correspond to your S3 client.

## Configure MinIO

If you are not using the MinIO client, skip this section.

To install and configure MinIO, complete the following steps:

1. Install the MinIO client on your platform as
described on the [MinIO download page](https://min.io/download#/macos).

1. Verify that MinIO components are successfully installed by running
the following command:

   ```bash
   minio version
   mc version
   ```

   **System Response:**

   ```
   Version: 2019-07-11T19:31:28Z
   Release-tag: RELEASE.2019-07-11T19-31-28Z
   Commit-id: 31e5ac02bdbdbaf20a87683925041f406307cfb9
   ```

1. Set up the MinIO configuration file to use the `30600` port for your host:

   ```bash
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

## Configure the AWS CLI

If you are not using the AWS CLI, skip this section.

If you have not done so already, you need to install and
configure the AWS CLI client on your machine. To configure the AWS CLI,
complete the following steps:

1. Install the AWS CLI for your operating system as described
in the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).

1. Verify that the AWS CLI is installed:

   ```bash
   aws --version
   ```

1. Configure AWS CLI:

   ```bash
   aws configure
   ```

   **System Response:**

   ```
   AWS Access Key ID: YOUR-PACHYDERM-AUTH-TOKEN
   AWS Secret Access Key: YOUR-PACHYDERM-AUTH-TOKEN
   Default region name:
   Default output format [None]:
   ```

   Both the access key and secret key should be set to your
   Pachyderm authentication token. If authentication is not enabled
   on the cluster, both parameters must be empty strings.

## Configure S3cmd

If you are not using S3cmd, skip this section.

S3cmd is an open-source command line client that enables you
to access S3 object store buckets. To configure S3cmd, complete
the following steps:
1. If you do not have S3cmd installed on your machine, install
it as described in the [S3cmd documentation](https://s3tools.org/download).
For example, in macOS, run:

   ```bash
   brew install s3cmd
   ```

1. Verify that S3cmd is installed:

   ```bash
   s3cmd --version
   s3cmd version 2.0.2
   ```

1. Configure S3cmd to use Pachyderm:

   ```bash
   s3cmd --configure
   ...
   ```

1. Fill all fields and specify the following settings for Pachyderm.

   **Example:**

   ```
   New settings:
   Access Key: "YOUR-PACHYDERM-AUTH-TOKEN"
   Secret Key: "YOUR-PACHYDERM-AUTH-TOKEN"
   Default Region: US
   S3 Endpoint: localhost:30600
   DNS-style bucket+hostname:port template for accessing a bucket: localhost:30600/%(bucket)
   Encryption password:
   Path to GPG program: /usr/local/bin/gpg
   Use HTTPS protocol: False
   HTTP Proxy server name:
   HTTP Proxy server port: 0
   ```

   Set the access key and secret key to your
   Pachyderm authentication token. If authentication is not
   enabled on the cluster, you can put any value.
