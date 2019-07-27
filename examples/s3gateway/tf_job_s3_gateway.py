from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
import os
import argparse
import sys
import tensorflow as tf
from tensorflow.python.lib.io import file_io


# this is a directory with four files in it,
# mf, sis, boom, and bah,
# that's built into the image
input_path  = os.getenv('INPUT_PATH', "/blah")
s3_bucket = os.getenv('S3_BUCKET', 'images.master')


def main(_):
    s3_path = 's3://' + args.bucket + '/'
    print("walking {}".format(args.inputpath))
    for dirpath, dirs, files in os.walk(args.inputpath):
      for file in files:
        print("copying {} to {} as {}".format(dirpath + "/" + file, s3_path, file))
        file_io.copy(dirpath + "/" + file, s3_path + file, True)
        print("printing {} as string: >>{}<<".format(s3_path + file, file_io.read_file_to_string(s3_path + file, False)))

if __name__ == '__main__':
  parser = argparse.ArgumentParser(description='put some files into an s3 bucket.')

  parser.add_argument('-b', '--bucket', required=False,
                      help="""The bucket where files will be put. This overrides the default and the environment variable S3_BUCKET.""",
                      default=s3_bucket)
  parser.add_argument('-i', '--inputpath', required=False,
                      help="""The directories to walk for files to put in the bucket""",
                      default=input_path)
  
  args, unparsed = parser.parse_known_args()
  tf.app.run(main=main, argv=[sys.argv[0]] + unparsed)

