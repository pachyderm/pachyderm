from __future__ import absolute_import, division, print_function, unicode_literals

import os
import argparse
import sys
import tensorflow as tf
from tensorflow import keras
import time
import numpy as np
import zipfile
import python_pachyderm
from tensorflow.python.lib.io import file_io
from tensorflow.python.keras.utils.data_utils import get_file


tf.__version__


# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket  = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir  = os.getenv('DATA_DIR', "/data")
# this is the training data file in the input repo
training_data = os.getenv('TRAINING_DATA', "mninst.npz")
# this is the name of model file in the output repo
model_file = os.getenv('MODEL_FILE', "my_model.h5")

client = python_pachyderm.PfsClient(host=os.getenv('PACHD_SERVICE_HOST'), port=os.getenv('PACHD_SERVICE_PORT'))

def main(_):
  input_url = 's3://' + args.inputbucket + "/"
  output_url = 's3://' + args.outputbucket + "/"
  
  os.makedirs(args.datadir)

  # first, we copy files from pachyderm into a convenient
  # local directory for processing.  
  input_uri = os.path.join(input_url, args.trainingdata)
  training_data_path = os.path.join(args.datadir, args.trainingdata)
  print("copying {} to {}".format(input_uri, training_data_path))
  file_io.copy(input_uri, training_data_path, True)
  
  (train_images, train_labels), (test_images, test_labels) = tf.keras.datasets.mnist.load_data(path=training_data_path)
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
  model_file =  os.path.join(args.datadir,args.modelfile)
  model.save(model_file)
  # Copy file over to Pachyderm
  output_uri = os.path.join(output_url,args.modelfile)
  print("copying {} to {}".format(model_file, output_uri))
  file_io.copy(model_file, output_uri, True)

  client.finish_commit(('output', 'master'))

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
  parser.add_argument('-t', '--trainingdata', required=False,
                      help="The training data used as input, in npz format.  This overrides the environment variable TRAINING_DATA. Default is {}".format(training_data),
                      default=training_data)
  parser.add_argument('-m', '--modelfile', required=False,
                      help="The filename of the model file to be output.  This overrides the environment variable MODEL_FILE. Default is {}".format(model_file),
                      default=model_file)
  
  args, unparsed = parser.parse_known_args()
  #tf.app.run(main=main, argv=[sys.argv[0]] + unparsed)
  tf.compat.v1.app.run(main=main, argv=[sys.argv[0]] + unparsed)
