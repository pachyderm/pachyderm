from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
import os
import argparse
import sys
import tensorflow as tf
from tensorflow.python.lib.io import file_io

# this is a directory with two directories in it
# containing the first chapters of Moby Dick and A Tale of Two Cities
# that's built into the image
input_path  = os.getenv('INPUT_PATH', "/testdata")
s3_bucket = os.getenv('S3_BUCKET', 'master.testrepo')


def main(_):
    s3_path = 's3://' + args.bucket
    print("walking {} for copying to {}".format(args.inputpath, s3_path))
    for dirpath, dirs, files in os.walk(args.inputpath, topdown=True):   
      for file in files:
        newpath = s3_path + dirpath + "/" +  file
        print("copying {} to {}".format(dirpath + "/" + file, newpath))
        file_io.copy(dirpath + "/" + file, newpath, True)

    targetpath = s3_path + args.inputpath
    print("walking {} for reading files copied".format(targetpath))
    for dirpath, dirs, files in file_io.walk(targetpath, True):
        for file in files:
            newpath = dirpath + "/" +  file
            print("printing {} in {} as string: >>{}<<".format(file, dirpath, file_io.read_file_to_string(newpath, False)))

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

# todo
# create image and place in pachyderm repo
# write README.md
# create PR
