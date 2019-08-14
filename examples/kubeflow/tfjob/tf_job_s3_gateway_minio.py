from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
# Import MinIO library.
from minio import Minio
from minio.error import (ResponseError, BucketAlreadyOwnedByYou,
                         BucketAlreadyExists)
import os
import argparse
import sys

import tensorflow as tf


unspecified_value = '';
s3_access_key = os.getenv('S3_ACCESS_KEY', unspecified_value)
s3_secret_key = os.getenv('S3_SECRET_KEY', unspecified_value)
s3_secure = os.getenv('S3_SECURE', 0)
s3_bucket = os.getenv('S3_BUCKET', 'master.testrepo')
s3_endpoint = os.getenv('S3_ENDPOINT', 'localhost:30060')
# this is a directory with four files in it,
# mf, sis, boom, and bah,
# that's built into the image
input_path  = os.getenv('INPUT_PATH', "/testdata")


minio_secure = False
if (s3_secure > 0):
    minio_secure = True
  

def main(_):
    # Initialize minioClient with an endpoint and access/secret keys.
    print("opening {}".format(args.endpoint))
    try:
      minioClient = Minio(args.endpoint,
                          access_key=args.accesskey,
                          secret_key=args.secretkey,
                          secure=args.secure)
    except Error as err:
      print (err)
    
    print("opened {}".format(minioClient))

    print("walking {}".format(args.inputpath))
    for dirpath, dirs, files in os.walk(args.inputpath):
        for file in files:
            try:
                print("copying {} to {} as {}".format(dirpath + "/" + file, args.bucket, file))
                minioClient.fput_object(args.bucket, file, dirpath + "/" + file)
            except ResponseError as err:
                print(err)
            print("copied {} to {} as {}".format(dirpath + "/" + file, args.bucket, file))




if __name__ == '__main__':
  parser = argparse.ArgumentParser(description='put some randomly generated files into an s3 bucket.')

  parser.add_argument('-b', '--bucket', required=False,
                      help="""The bucket where files will be put. This overrides the default and the environment variable S3_BUCKET.""",
                      default=s3_bucket)
  parser.add_argument('-e', '--endpoint', required=False,
                      help="""S3 endpoint, hostname:port. This overrides the default and the environment variable S3_ENDPOINT.""",
                      default=s3_endpoint)
  parser.add_argument('-s', '--secure', required=False,
                      help="""Whether the S3 endpoint is using https.  This overrides the default and the environment variable S3_SECURE.""",
                      default=minio_secure)
  parser.add_argument('-a', '--accesskey', required=False,
                      help="""Access key for the bucket. This overrides the default and the environment variable S3_SECURE.""",
                      default=s3_access_key)
  parser.add_argument('-k', '--secretkey', required=False,
                      help="""Secret key for the bucket.  This overrides the default and the environment variable S3_SECRET_KEY.""",
                      default=s3_secret_key)
  parser.add_argument('-i', '--inputpath', required=False,
                      help="""The directories to walk for files to put in the bucket""",
                      default=input_path)
  
  args, unparsed = parser.parse_known_args()
  tf.app.run(main=main, argv=[sys.argv[0]] + unparsed)

