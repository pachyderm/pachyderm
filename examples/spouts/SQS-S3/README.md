# Amazon SQS-S3 Spout

This is a simple spout that listens for "object added" notifications on an SQS topic, grabs the files, and places them in its output repo.

To use the SQS spout, complete the following steps:

1. Create an S3 Bucket.
2. Create an SQS topic that subscribes to all create-object events in the bucket.
3. Create a Pachyderm spout with appropriate environment variables or parameters to access the bucket and topic.


Usage:

```
sqs-spout.py [-h] [-i AWS_ACCESS_KEY_ID] [-k AWS_SECRET_ACCESS_KEY]
                    [-r AWS_REGION] [-o OUTPUT_PIPE] [-b S3_BUCKET]
                    [-q SQS_QUEUE_URL]

Listen on an SQS queue for notifications of added files to an S3 bucket and
then ingress the files.

optional arguments:
  -h, --help            show this help message and exit
  -i AWS_ACCESS_KEY_ID, --aws_access_key_id AWS_ACCESS_KEY_ID
                        AWS Access Key ID for accessing the SQS queue and the
                        bucket. Overrides env var AWS_ACCESS_KEY_ID. Default:
                        'user-id'.
  -k AWS_SECRET_ACCESS_KEY, --aws_secret_access_key AWS_SECRET_ACCESS_KEY
                        AWS secret key for accessing the SQS queue and the
                        bucket. Overrides env var AWS_SECRET_ACCESS_KEY.
                        Default: 'secret-key'.
  -r AWS_REGION, --aws_region AWS_REGION
                        AWS region. Overrides env var AWS_REGION. Default:
                        'us-east-1'
  -o OUTPUT_PIPE, --output_pipe OUTPUT_PIPE
                        The named pipe that the tarstream containing the files
                        will be written to. Overrides env var OUTPUT_PIPE.
                        Default: '/pfs/out'.
  -b S3_BUCKET, --s3_bucket S3_BUCKET
                        The url to the SQS queue for bucket notifications.
                        Overrides env var S3_BUCKET. Default: 's3://bucket-
                        name/'.
  -q SQS_QUEUE_URL, --sqs_queue_url SQS_QUEUE_URL
                        The url to the SQS queue for bucket notifications.
                        Overrides env var SQS_QUEUE_URL. Default:
                        'https://sqs.us-west-1.amazonaws.com/ID/Name'.

```
