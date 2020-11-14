# Amazon SQS S3 Spout

This example describes how to create a simple spout
that listens for "object added" notifications on an
Amazon™ Simple Queue Service (SQS) queue, grabs the
files, and places them into a Pachyderm repository.

## Prerequisites

You must have the following configured in your environment to
run this example:

* An AWS account
* Pachyderm 1.9.5 or later

## Configure AWS Prerequisites

Before you can run this spout, you need to configure
an S3 bucket, a Simple Notification Service (SNS),
and an SQS queue in your AWS account.

Complete the following steps:

1. Create an S3 Bucket.
2. Create an SNS topic and an SQS queue as described in
the [Amazon Documentation](https://docs.aws.amazon.com/AmazonS3/latest/dev/ways-to-add-notification-config-to-bucket.html).
3. In your S3 bucket, add an event notification:

   1. Select your S3 bucket.
   2. Go to **Properties**.
   3. Click **Events > Add notification**.
   4. Select **All object create events**.
   5. In **Send to**, select **SQS Queue** and pick your
   SQS queue from the dropdown list.

4. Test that the SNS topic and SQS are working by adding a test
   file into your S3 bucket. You should get an email notification
   about a new object created in the bucket.

## Create a Spout

Use [the SQS example pipeline specification](sqs-spout.json)
and [the sample Python script](sqs-spout.py)
to create a spout pipeline:

1. Clone the Pachyderm repository:

   ```bash
   $ git clone git@github.com:pachyderm/pachyderm.git
   ```

1. Add the following environment variables to `sqs-spout.json`:

   * `AWS_REGION`
   * `OUTPUT_PIPE`
   * `S3_BUCKET`
   * `SQS_QUEUE_URL`
   * `VERBOSE_LOGGING`

   For more information, see [Pipeline Environment Parameters](#pipeline-environment-parameters).

1. Add a secret with the following two keys

   * `AWS_ACCESS_KEY_ID`
   * `AWS_SECRET_ACCESS_KEY`

   The values `<your-password>` and `<account name>` are enclosed in single quotes to prevent the shell from interpreting them.
   
   ```sh
   $ echo -n '<account-name>' > AWS_ACCESS_KEY_ID ; chmod 600 AWS_ACCESS_KEY_ID
   $ echo -n '<your-password>' > AWS_SECRET_ACCESS_KEY ; chmod 600 AWS_SECRET_ACCESS_KEY
   ```
   
1. Confirm the values in these files are what you expect.

   ```sh
   $ cat AWS_ACCESS_KEY_ID
   $ cat AWS_SECRET_ACCESS_KEY
   ```
   
   The output from those two commands should be `<account-name>` and `<your-password>`, respectively.
   
   Creating the secret will require different steps,
   depending on whether you have Kubernetes access or not.
   Pachyderm Hub users don't have access to Kubernetes.
   If you have Kubernetes access, 
   follow the two steps prefixed with "(Kubernetes)".
   If you don't have access to Kubernetes,
   follow the three steps labeled "(Pachyderm Hub)" 

1. (Kubernetes) If you have direct access to the Kubernetes cluster, you can create a secret using `kubectl`.
   
   ```sh
   $ kubectl create secret generic aws-credentials --from-file=./AWS_ACCESS_KEY_ID --from-file=./AWS_SECRET_ACCESS_KEY
   ```
   
1. (Kubernetes) Confirm that the secrets got set correctly.
   You use `kubectl get secret` to output the secrets, and then decode them using `jq` to confirm they're correct.
   
   ```sh
   $ kubectl get secret aws-credentials -o json | jq '.data | map_values(@base64d)'
   {
       "AWS_ACCESS_KEY_ID": "<account-name>",
       "AWS_SECRET_ACCESS_KEY": "<your-password>"
   }
   ```

   You will have to use pachctl if you're using Pachyderm Hub,
   or don't have access to the Kubernetes cluster.
   The next three steps show how to do that.

1. (Pachyderm Hub) Create a secrets file from the provided template.

   ```sh
   $ jq -n --arg AWS_ACCESS_KEY_ID $(cat AWS_ACCESS_KEY_ID) --arg AWS_SECRET_ACCESS_KEY $(cat AWS_SECRET_ACCESS_KEY) \
         -f aws-credentials-template.jq  > aws-credentials-secret.json 
   $ chmod 600 aws-credentials-secret.json
   ```

1. (Pachyderm Hub) Confirm the secrets file is correct by decoding the values.

   ```sh
   $ jq '.data | map_values(@base64d)' aws-credentials-secret.json
   {
       "AWS_ACCESS_KEY_ID": "<account-name>",
       "AWS_SECRET_ACCESS_KEY": "<your-password>"
   }
   ```

1. (Pachyderm Hub) Generate a secret using pachctl

   ```sh
   $ pachctl create secret -f mongodb-credentials-secret.json
   ```
   
1. Create a pipeline from `sqs-spout.json`:

   ```bash
   $ pachctl create pipeline -f sqs-spout.json
   ```

1. Verify that the pipeline was created:

   ```bash
   $ pachctl list pipeline
   NAME       VERSION INPUT    CREATED        STATE / LAST JOB
   sqs-spout  1       none     2 minutes ago  running / starting
   ```

   You should also see that an output repository was created for your
   spout pipeline:

   ```bash
   $ pachctl list repo
   NAME       CREATED       SIZE
   sqs-spout  2 minutes ago 0B
   ```

## Run the Spout

After you create an SQS spout, you can test it by uploading a file
into your S3 bucket and later finding it in the
SQS pipeline output repository.

To test the spout, complete the following steps:

1. In the IAM console, go to S3 and find your bucket.

1. Upload a file into your bucket. For example, `01-pipeline.png`. Depending
on the size of the file, it might take some time for the file to get uploaded.

1. In your terminal, run:

   ```bash
   $ pachctl list commit sqs-spout
   REPO      BRANCH COMMIT                           PARENT    STARTED        DURATION           SIZE
   sqs-spout master 4ecc933d523d485b8a9cce6b1feeac95 none      6 minutes ago  Less than a second 37.44KiB
   ```

1. Verify that the file that you have uploaded to the S3 bucket is
in the `sqs-spout` output repository. Example:

   ```bash
   $ pachctl list file sqs-spout@master
   NAME             TYPE SIZE
   /01-pipeline.png file 37.44KiB
   ```

## Pipeline Environment Parameters

This table describes pipeline parameters that you can specify in your
pipeline specification.

| Optional Parameter  | Description   |
| ------------------- | ------------- |
| `-i AWS_ACCESS_KEY_ID`, `--aws_access_key_id AWS_ACCESS_KEY_ID` | An AWS Access Key ID for accessing the SQS queue and the bucket. Overrides env var AWS_ACCESS_KEY_ID. The default value is `user-id`. You can view your AWS credentials in your AWS Management Console or, if you have set up AWS CLI, in the `~/.aws/config` file. |
| `-k AWS_SECRET_ACCESS_KEY`, `--aws_secret_access_key AWS_SECRET_ACCESS_KEY` | AWS secret key for accessing the SQS queue and the bucket. Overrides env var AWS_SECRET_ACCESS_KEY. The default value is `secret-key`. You can view your AWS credentials in your AWS Management Console or, if you have set up AWS CLI, in the `~/.aws/config` file. |
| `-r AWS_REGION`, `--aws_region AWS_REGION` | An AWS region. Overrides env var `AWS_REGION`. The default value is `us-east-1`. |
| `-o OUTPUT_PIPE`, `--output_pipe OUTPUT_PIPE` | The named pipe that the tar stream that contains the files is written to. Overrides env var `OUTPUT_PIPE`. The default value is `/pfs/out`. |
| `-b S3_BUCKET`, `--s3_bucket S3_BUCKET` | The URL to the SQS queue for bucket notifications. Overrides env var `S3_BUCKET`. The default values is `s3://bucket-name/`. |
| `-q SQS_QUEUE_URL`, `--sqs_queue_url SQS_QUEUE_URL` | The URL to the SQS queue for bucket notifications. Overrides env var `SQS_QUEUE_URL`. The default value is `https://sqs.us-west-1.amazonaws.com/ID/Name`. |
