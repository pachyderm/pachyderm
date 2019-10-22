import os
import json
import logging
import boto3
import tarfile
import gzip
import stat
import io
import time
import argparse
from datetime import datetime
from botocore.exceptions import ClientError
from pprint import pprint


aws_access_key_id = os.getenv('AWS_ACCESS_KEY_ID', 'user-id')
aws_secret_access_key = os.getenv('AWS_SECRET_ACCESS_KEY', 'secret-key')
aws_region = os.getenv('AWS_REGION', 'us-east-1')
output_pipe = os.getenv('OUTPUT_PIPE', '/pfs/out')
sqs_queue_url = os.getenv('SQS_QUEUE_URL', 'https://sqs.us-west-1.amazonaws.com/ID/Name')
s3_bucket = os.getenv('S3_BUCKET', 's3://bucket-name/')

def retrieve_sqs_messages(sqs_queue_url, aws_access_key_id, aws_secret_access_key, aws_region, num_msgs=1, wait_time=0, visibility_time=5):
    """Retrieve messages from an SQS queue

    The retrieved messages are not deleted from the queue.

    :param sqs_queue_url: String URL of existing SQS queue
    :param num_msgs: Number of messages to retrieve (1-10)
    :param wait_time: Number of seconds to wait if no messages in queue
    :param visibility_time: Number of seconds to make retrieved messages
        hidden from subsequent retrieval requests
    :return: List of retrieved messages. If no messages are available, returned
        list is empty. If error, returns None.
    """

    # Validate number of messages to retrieve
    if num_msgs < 1:
        num_msgs = 1
    elif num_msgs > 10:
        num_msgs = 10

    # Retrieve messages from an SQS queue
    sqs_client = boto3.client('sqs', aws_access_key_id=aws_access_key_id, aws_secret_access_key=aws_secret_access_key, region_name=aws_region)
    try:
        msgs = sqs_client.receive_message(QueueUrl=sqs_queue_url,
                                          MaxNumberOfMessages=num_msgs,
                                          WaitTimeSeconds=wait_time,
                                          VisibilityTimeout=visibility_time)
    except ClientError as e:
        logging.error(e)
        return None

    # Return the list of retrieved messages
    return msgs['Messages']


def delete_sqs_message(sqs_queue_url, msg_receipt_handle):
    """Delete a message from an SQS queue

    :param sqs_queue_url: String URL of existing SQS queue
    :param msg_receipt_handle: Receipt handle value of retrieved message
    """

    # Delete the message from the SQS queue
    sqs_client = boto3.client('sqs')
    sqs_client.delete_message(QueueUrl=sqs_queue_url,
                              ReceiptHandle=msg_receipt_handle)

def open_pipe(path_to_file, attempts=0, timeout=2, sleep_int=5):
    if attempts < timeout : 
        flags = os.O_WRONLY  # Refer to "man 2 open".
        mode = stat.S_IWUSR  # This is 0o400.
        umask = 0o777 ^ mode  # Prevents always downgrading umask to 0.
        umask_original = os.umask(umask)
        try:
            file = os.open(path_to_file, flags, mode)
            # you must open the pipe as binary to prevent line-buffering problems.
            return os.fdopen(file, "wb")
        except OSError as oe:
            print ('{0} attempt of {1}; error opening file: {2}'.format(attempts + 1, timeout, oe))
            os.umask(umask_original)
            time.sleep(sleep_int)
            return open_pipe(path_to_file, attempts + 1)
        finally:
            os.umask(umask_original)
    return None

def main():
    parser = argparse.ArgumentParser(description='Listen on an SQS queue for notifications of added files to an S3 bucket and then ingress the files.')

    parser.add_argument('-i', '--aws_access_key_id', required=False,
                        help="AWS Access Key ID for accessing the SQS queue and the bucket. Overrides env var AWS_ACCESS_KEY_ID. Default: '{0}'.".format(aws_access_key_id),
                        default=aws_access_key_id)
    parser.add_argument('-k', '--aws_secret_access_key', required=False,
                        help="AWS secret key for accessing the SQS queue and the bucket. Overrides env var AWS_SECRET_ACCESS_KEY. Default: '{0}'.".format(aws_secret_access_key),
                        default=aws_secret_access_key)
    parser.add_argument('-r', '--aws_region', required=False,
                        help="AWS region.  Overrides env var AWS_REGION. Default: '{0}'".format(aws_region),
                        default=aws_region)
    parser.add_argument('-o', '--output_pipe', required=False,
                        help="The named pipe that the tarstream containing the files will be written to.  Overrides env var OUTPUT_PIPE. Default: '{0}'.".format(output_pipe),
                        default=output_pipe)
    parser.add_argument('-b', '--s3_bucket', required=False,
                        help="The url to the SQS queue for bucket notifications.  Overrides env var S3_BUCKET. Default: '{0}'.".format(s3_bucket),
                        default=s3_bucket)
    parser.add_argument('-q', '--sqs_queue_url', required=False,
                        help="The url to the SQS queue for bucket notifications.  Overrides env var SQS_QUEUE_URL. Default: '{0}'.".format(sqs_queue_url),
                        default=sqs_queue_url)

    
    args = parser.parse_args()

    num_messages = 1
    s3 = boto3.client('s3')

    # Set up logging
    logging.basicConfig(level=logging.DEBUG,
                        format='%(levelname)s: %(asctime)s: %(message)s')

    # Retrieve SQS messages
    msgs = retrieve_sqs_messages(args.sqs_queue_url, args.aws_access_key_id, args.aws_secret_access_key, args.aws_region, num_messages)
    if msgs is not None:
        for msg in msgs:
            logging.info(f'SQS: Message ID: {msg["MessageId"]}, '
                         f'Contents: {msg["Body"]}')
            # Iterate over the JSON to pull out full file name 
            data = json.loads(msg['Body'])['Records'][0]
            bucket = data['s3']['bucket']['name']
            file = data['s3']['object']['key']

            # Get the file from S3 and extract it.
            file_path = os.path.join(args.s3_bucket,file)
            print(file_path)
            s3.download_file(bucket, file, file)
           
            
            # Start Spout portion
            mySpout = open_pipe(args.output_pipe)

            # To use a tarfile object with a named pipe, you must use the "w|" mode
            # which makes it not seekable
            print("Creating tarstream...")
            try: 
                tarStream = tarfile.open(fileobj=mySpout,mode="w|", encoding='utf-8')
            except tarfile.TarError as te:
                print('error creating tarstream: {0}'.format(te))
                exit(-2)

            print("Creating tar archive entry file {}...".format(file))
            
            tarHeader = tarfile.TarInfo(file)
            tarHeader.size = os.stat(file).st_size #Size of the file itself
            tarHeader.mode = 0o600
            tarHeader.name = file
            print(tarHeader.size)

            print("Writing tarfile to spout for file: {}...".format(file))
            try:
                with open(file, mode="rb") as file:
                    tarStream.addfile(tarinfo=tarHeader, fileobj=file)
            except tarfile.TarError as te:
                print('error writing message {0} to tarstream: {1}'.format(file, te))
                exit(-2)
            tarStream.close()
            mySpout.close()

            delete_sqs_message(sqs_queue_url, msg['ReceiptHandle'])

if __name__ == '__main__':
    main()
