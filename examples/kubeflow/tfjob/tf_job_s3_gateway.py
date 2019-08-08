from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
import os
import argparse
import sys
import tensorflow as tf
from tensorflow.python.lib.io import file_io

# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket  = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir  = os.getenv('DATA_DIR', "/data")


def main(_):

    # The Tensorflow file_io.walk() function has an issue
    # with iterating over the top level of a bucket.
    # It requires a directory within the bucket.
    # So, we give it one.
    input_url = 's3://' + args.inputbucket + "/data/"
    output_url = 's3://' + args.outputbucket + "/data/"

    os.makedirs(args.datadir)

    # first, we copy files from pachyderm into a convenient
    # local directory for processing.  The files have been
    # placed into the inputpath directory in the s3path bucket.
    print("walking {} for copying files".format(input_url))
    for dirpath, dirs, files in file_io.walk(input_url, True):
        for file in files:
            uri = os.path.join(dirpath, file)
            newpath = os.path.join(args.datadir, file)
            print("copying {} to {}".format(uri, newpath))
            file_io.copy(uri, newpath, True)


    # here is where you would apply your training to the data in args.datadir
    # it might operate on the data directly, or place additional
    # data in the same directory

    # finally, we copy the output from those operations to
    # another pachyderm repo
    print("walking {} for copying to {}".format(args.datadir, output_url))
    for dirpath, dirs, files in os.walk(args.datadir, topdown=True):   
      for file in files:
        uri = os.path.join(dirpath, file)
        newpath = output_url + file
        print("copying {} to {}".format(uri, newpath))
        file_io.copy(uri, newpath, True)

if __name__ == '__main__':
  parser = argparse.ArgumentParser(description='Copy data from an S3 input bucket, operate on it, and copy the data to a different S3 bucket.')

  parser.add_argument('-i', '--inputbucket', required=False,
                      help="The bucket where files will be copied from. This overrides the environment variable INPUT_BUCKET. Default is {}".format(input_bucket),
                      default=input_bucket)
  parser.add_argument('-o', '--outputbucket', required=False,
                      help="The bucket where results will be copied to. This overrides the environment variable OUTPUT_BUCKET. Default is {}".format(output_bucket),
                      default=output_bucket)
  parser.add_argument('-d', '--datadir', required=False,
                      help="The local directory where data will be copied to.  This overrides the environment variable DATA_DIR. Default is {}".format(data_dir),
                      default=data_dir)
  
  args, unparsed = parser.parse_known_args()
  tf.app.run(main=main, argv=[sys.argv[0]] + unparsed)

