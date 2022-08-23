# Configure Your S3 Client

Before you can access a repo via the S3 gateway,
you need to configure your S3 client. 
Complete the steps in one of the sections below that
correspond to your S3 client.

## Configure MinIO
1. Install the MinIO client as
described on the [MinIO download page](https://min.io/download){target=_blank}.

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
      You should see a configuration similar to the following.
      For a minikube deployment, verify the
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

      Both the access key and secret key 
      should be set as mentioned in the [# Set Your Credentails](#set-your-credentials) section of this page. 

!!! Example "Example:  Check the list of filesystem objects on the `master` branch of the repository `raw_data`"
      ```shell
      mc ls local/master.raw_data
      ```

!!! Info
      Find **MinIO** full documentation [here](https://docs.min.io/docs/minio-client-complete-guide){target=_blank}.

## Configure The AWS CLI
1. Install the AWS CLI as described
in the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html){target=_blank}.

1. Verify that the AWS CLI is installed:

      ```shell
      aws --version
      ```

1. Configure AWS CLI. Use the `aws configure` command to configure your credentials file:
      ```shell
      aws configure --profile <name-your-profile>
      ```
      Both the access key and secret key 
      should be set as mentioned in the [# Set Your Credentails](#set-your-credentials) section of this page.

      **System Response:**
      ```
      AWS Access Key ID: YOUR-PACHYDERM-AUTH-TOKEN
      AWS Secret Access Key: YOUR-PACHYDERM-AUTH-TOKEN
      Default region name:
      Default output format [None]:
      ```
!!! Note
      Note that the `--profile` flag ([named profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html){target=_blank}) is optional. If not used, your access information will be stored in the default profile. 
      
      To reference a given profile when using the S3 client, append `--profile <name-your-profile>` at the end of your command.

!!! Example "Example:  Check the list of filesystem objects on the `master` branch of the repository `raw_data`"
      ```shell
      aws --endpoint-url http://<localhost_or_externalIP>:30600/ s3 ls s3://master.raw_data
      ```

!!! Info
      Find **AWS S3 CLI** full documentation [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-services-s3-commands.html){target=_blank}.
 
## Configure boto3
Before using Boto3, you need to [set up authentication credentials for your AWS account](#configure-the-aws-cli) using the AWS CLI as mentioned previously.

Then follow the [Using boto](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/quickstart.html#using-boto3){target=_blank} documentation starting with importing boto3 in your python file and creating your S3 resources.
   
!!! Info   
      Find **boto3** full documentation [here](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html){target=_blank}.


## Set Your Credentials
- If [authentication is enabled](../../../../enterprise/auth/), 
retrieve your session token in your active context:

      ```shell
      more ~/.pachyderm/config.json
      ```
      Search for your session token: `"session_token": "your-session-token-value"`.
      **Make sure to fill both fields `Access Key ID` and `Secret Access Key` with that same value.**

      Depending on your use case, it might make sense to pass the credentials of a robot-user or another type of user altogether. Refer to the [authentication section of the documentation](../../../../enterprise/auth/authorization/) for more RBAC information.

- If the authentication feature is not activated, make sure that whether you fill in those fields or not, their content always matches. (i.e., both empty or both set to the same value)

## Next

Check all the [operations supported by Pachyderm's S3 gateway](../supported-operations)

