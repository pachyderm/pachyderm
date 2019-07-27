# Copyright 2015 The TensorFlow Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the 'License');
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an 'AS IS' BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
"""A simple MNIST classifier which displays summaries in TensorBoard.
This is an unimpressive MNIST model, but it is a good example of using
tf.name_scope to make a graph legible in the TensorBoard graph explorer, and of
naming summary tags so that they are grouped meaningfully in TensorBoard.
It demonstrates the functionality of every TensorBoard dashboard.
"""
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
s3_bucket = os.getenv('S3_BUCKET', 'images.master')
s3_endpoint = os.getenv('S3_ENDPOINT', 'localhost:30060')
# this is a directory with four files in it,
# mf, sis, boom, and bah,
# that's built into the image
input_path  = os.getenv('INPUT_PATH', "/blah")


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

