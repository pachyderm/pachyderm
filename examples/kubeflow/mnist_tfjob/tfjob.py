from __future__ import absolute_import, division, print_function, unicode_literals

import os
import argparse
import sys
import tensorflow as tf
from tensorflow import keras
import time
import numpy as np
import zipfile
from tensorflow.python.lib.io import file_io
from tensorflow.python.keras.utils.data_utils import get_file


tf.__version__


# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket  = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir  = os.getenv('DATA_DIR', "/data")

# Getting the Pachyderm stuff stared
#input_url = 's3://' + args.inputbucket + "/data/"
#output_url = 's3://' + args.outputbucket + "/data/"
def main(_):
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

  path = '/tmp/data/mnist.npz'
  (train_images, train_labels), (test_images, test_labels) = tf.keras.datasets.mnist.load_data(path)

  train_labels = train_labels[:1000]
  test_labels = test_labels[:1000]

  train_images = train_images[:1000].reshape(-1, 28 * 28) / 255.0
  test_images = test_images[:1000].reshape(-1, 28 * 28) / 255.0

  # Returns a short sequential model
  def create_model():
    model = tf.keras.models.Sequential([
      keras.layers.Dense(512, activation=tf.keras.activations.relu, input_shape=(784,)),
      keras.layers.Dropout(0.2),
      keras.layers.Dense(10, activation=tf.keras.activations.softmax)
      ])

    model.compile(optimizer=tf.keras.optimizers.Adam(),
      loss=tf.keras.losses.sparse_categorical_crossentropy,
      metrics=['accuracy'])

    return model

  # Create a basic model instance
  model = create_model()
  model.summary()
  

  model.fit(train_images, train_labels, batch_size=32, epochs=5,
            validation_data=(test_images, test_labels))

  # Save entire model to a HDF5 file
  os.mkdir('/tmp/data/models')
  model.save('/tmp/data/models/my_model.h5')

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
  #tf.app.run(main=main, argv=[sys.argv[0]] + unparsed)
  tf.compat.v1.app.run(main=main, argv=[sys.argv[0]] + unparsed)