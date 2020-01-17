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
logging_verbosity = os.getenv('LOGGING_VERBOSITY', "critical")
logging_boto = os.getenv('LOGGING_BOTO', "critical")
timeout = os.getenv('TIMEOUT', 5)
logging_format = '%(levelname)s: %(asctime)s: %(message)s'

def retrieve_sqs_messages(sqs_client, sqs_queue_url,  num_msgs=1, wait_time=0, visibility_time=5):
    """Retrieve messages from an SQS queue

    The retrieved messages are not deleted from the queue.

    :param sqs_queue_url: String URL of existing SQS queue
    :param aws_access_key_id: AWS user with access to queue
    :param aws_secret_access_key: AWS user secret key / password to access queue
    :param aws_region: region in which the queue resides
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
    try:
        msgs = sqs_client.receive_message(QueueUrl=sqs_queue_url,
                                          MaxNumberOfMessages=num_msgs,
                                          WaitTimeSeconds=wait_time,
                                          VisibilityTimeout=visibility_time)
    except ClientError as e:
        logging.error(e)
        return None

    logging.debug("returning msgs: {0}".format(msgs))
    # Return the list of retrieved messages
    if msgs and 'Messages' in msgs.keys():
        return msgs['Messages']

    return None


def delete_sqs_message(sqs_client, sqs_queue_url, msg_receipt_handle):
    """Delete a message from an SQS queue

    :param sqs_queue_url: String URL of existing SQS queue
    :param msg_receipt_handle: Receipt handle value of retrieved message
    """

    # Delete the message from the SQS queue
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
    parser.add_argument('-v', '--logging_verbosity', required=False,
                        help="Verbosity for logging: debug, info, warning, error, critical.  Overrides env var LOGGING_VERBOSITY. Default: '{0}'.".format(logging_verbosity),
                        default=logging_verbosity)
    parser.add_argument('-l', '--logging_boto', required=False,
                        help="Verbosity for logging boto3: debug, info, warning, error, critical.  Overrides env var LOGGING_BOTO. Default: '{0}'.".format(logging_boto),
                        default=logging_boto)
    parser.add_argument('-t', '--timeout', required=False, type=int,
                        help="Timeout for polling the SQS topic.  Overrides env var TIMEOUT. Default: '{0}'.".format(timeout),
                        default=timeout)

    
    args = parser.parse_args()

    num_messages = 1
    sqs_client = boto3.client('sqs', aws_access_key_id=args.aws_access_key_id, aws_secret_access_key=args.aws_secret_access_key, region_name=args.aws_region)
    s3_client = boto3.client('s3', aws_access_key_id=args.aws_access_key_id, aws_secret_access_key=args.aws_secret_access_key, region_name=args.aws_region)

    # Set up logging
    if args.logging_verbosity == "debug":
        logging.basicConfig(level=logging.DEBUG,
                            format=logging_verbosity)
    elif args.logging_verbosity == "info":
        logging.basicConfig(level=logging.INFO,
                            format=logging_format)
    elif args.logging_verbosity == "warning":
        logging.basicConfig(level=logging.WARNING,
                            format=logging_format)
    elif args.logging_verbosity == "error":
        logging.basicConfig(level=logging.ERROR,
                            format=logging_format)
    else:
        logging.basicConfig(level=logging.CRITICAL,
                            format=logging_format)

    # Set up logging for boto
    if args.logging_boto == "debug":
        logging.getLogger('boto3').setLevel(logging.DEBUG)
        logging.getLogger('botocore').setLevel(logging.DEBUG)
        logging.getLogger('nose').setLevel(logging.DEBUG)
    elif args.logging_boto == "info":
        logging.getLogger('boto3').setLevel(logging.INFO)
        logging.getLogger('botocore').setLevel(logging.INFO)
        logging.getLogger('nose').setLevel(logging.INFO)
    elif args.logging_boto == "warning":
        logging.getLogger('boto3').setLevel(logging.WARNING)
        logging.getLogger('botocore').setLevel(logging.WARNING)
        logging.getLogger('nose').setLevel(logging.WARNING)
    elif args.logging_boto == "error":
        logging.getLogger('boto3').setLevel(logging.ERROR)
        logging.getLogger('botocore').setLevel(logging.ERROR)
        logging.getLogger('nose').setLevel(logging.ERROR)
    else:
        logging.getLogger('boto3').setLevel(logging.CRITICAL)
        logging.getLogger('botocore').setLevel(logging.CRITICAL)
        logging.getLogger('nose').setLevel(logging.CRITICAL) 

    console = logging.StreamHandler()
    logging.getLogger('').addHandler(console)
    

    # Retrieve SQS messages
    logging.debug("starting sqs retrieval loop")
    while True: 
        logging.debug("retrieving {4} messages from {0} using access key starting with {1} and secret starting with {2} in region {3}.".format(
          args.sqs_queue_url, args.aws_access_key_id[0:5], args.aws_secret_access_key[0:5], args.aws_region, num_messages))
        msgs = retrieve_sqs_messages(sqs_client, args.sqs_queue_url, num_messages, wait_time=args.timeout)
        logging.debug("message: {0}.".format(msgs))
        if msgs is not None:
            for msg in msgs:
                logging.debug(f'SQS: Message ID: {msg["MessageId"]}, '
                             f'Contents: {msg["Body"]}')
                # Iterate over the JSON to pull out full file name 
            data = json.loads(msg['Body'])['Records'][0]
            bucket = data['s3']['bucket']['name']
            file = data['s3']['object']['key']

            # Get the file from S3 and extract it.
            file_path = os.path.join(args.s3_bucket,file)
            logging.debug("fetching url from s3: {0}".format(file_path))
            s3_client.download_file(bucket, file, file)
            
            
            # Start Spout portion
            logging.debug("opening pipe {0}".format(args.output_pipe))
            mySpout = open_pipe(args.output_pipe)

            # To use a tarfile object with a named pipe, you must use the "w|" mode
            # which makes it not seekable
            logging.debug("creating tarstream")
            try: 
                tarStream = tarfile.open(fileobj=mySpout,mode="w|", encoding='utf-8')
            except tarfile.TarError as te:
                logging.critical('error creating tarstream: {0}'.format(te))
                exit(-2)

            size = os.stat(file).st_size 
            logging.debug("Creating tar archive entry for file {0} of size {1}...".format(file, size))
            
            tarHeader = tarfile.TarInfo(file)
            tarHeader.size = size #Size of the file itself
            tarHeader.mode = 0o600
            tarHeader.name = file

            logging.debug("Writing tarfile to spout for file {0}...".format(file))
            try:
                with open(file, mode="rb") as file:
                    tarStream.addfile(tarinfo=tarHeader, fileobj=file)
            except tarfile.TarError as te:
                logging.critical('error writing message {0} to tarstream: {1}'.format(file, te))
                tarStream.close()
                mySpout.close()
                exit(-2)
                
            tarStream.close()
            mySpout.close()
            delete_sqs_message(sqs_client, args.sqs_queue_url, msg['ReceiptHandle'])

if __name__ == '__main__':
    main()
