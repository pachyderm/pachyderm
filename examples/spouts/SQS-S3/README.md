# Amazon SQS S3 Spout

This example describes how to create a simple spout
that listens for "object added" notifications on an
Amazon™ Simple Queue Service (SQS) queue, grabs the
files, and places them in its output repo.

## Configure AWS Prerequisites

Before you can run this spout, you need to configure
an S3 bucket, a Simple Notification Service (SNS),
and an SQS queue in your AWS account.

Complete the following steps:

1. Create an S3 Bucket.
2. Create an SNS topic and an SQS queue that subscribes
to all *create-object* events in the bucket. See
the [Amazon Documentation](https://docs.aws.amazon.com/AmazonS3/latest/dev/ways-to-add-notification-config-to-bucket.html).

## Create a Spout

Use [the SQS example pipeline specification](sqs-spout.json)
and [the sample Python script](sqs-spout.py)
to create a spout pipeline:

1. Clone the Pachyderm repository:

   ```bash
   $ git clone git@github.com:pachyderm/pachyderm.git
   ```

1. Add the environment variable to the `sqs-spout.json`.
1. Create a pipeline from `sqs-spout.json`:

   ```bash
   $ pachctl create pipeline -f /path/to/sqs-spout.json
   ```

   Verify that the pipeline was created:

   ```bash
   NAME       VERSION INPUT    CREATED        STATE / LAST JOB
   sqs-spout  1       none     2 minutes ago  running / starting
   ```

## Pipeline Environment Parameters

This table describes pipeline parameters that you can specify in your
pipeline specification.

| Optional Parameter  | Description   |
| ------------------- | ------------- |
| `-h`, `--help`          | Show help message and exit. |
| `-i AWS_ACCESS_KEY_ID`, `--aws_access_key_id AWS_ACCESS_KEY_ID` | An AWS Access Key ID for accessing the SQS queue and the bucket. Overrides env var AWS_ACCESS_KEY_ID. Default value is `user-id`. You can view your AWS credentials in your Amazon IAM Console or, if you have set up aws CLI in the `~/.aws/config` file. |
| `-k` AWS_SECRET_ACCESS_KEY`, `--aws_secret_access_key AWS_SECRET_ACCESS_KEY` | AWS secret key for accessing the SQS queue and the bucket. Overrides env var AWS_SECRET_ACCESS_KEY. Default value is `secret-key`. You can view your AWS credentials in your Amazon IAM Console or, if you have set up aws CLI in the `~/.aws/config` file. |
| `-r AWS_REGION`, `--aws_region AWS_REGION` | An AWS region. Overrides env var `AWS_REGION`. Default value is `us-east-1`. |
| `-o OUTPUT_PIPE`, `--output_pipe OUTPUT_PIPE` | The named pipe that the tarstream containing the files is written to. Overrides env var `OUTPUT_PIPE`.Default value is `/pfs/out`. |
| `-b S3_BUCKET`, `--s3_bucket S3_BUCKET` | The URL to the SQS queue for bucket notifications. Overrides env var `S3_BUCKET`. Default values is `s3://bucket-name/`. |
| `-q SQS_QUEUE_URL`, `--sqs_queue_url SQS_QUEUE_URL` | The URL to the SQS queue for bucket notifications. Overrides env var `SQS_QUEUE_URL`. Default value is `https://sqs.us-west-1.amazonaws.com/ID/Name`. |
