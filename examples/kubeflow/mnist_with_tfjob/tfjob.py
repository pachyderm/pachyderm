from __future__ import absolute_import, division, print_function, unicode_literals

import os
import sys
import tensorflow as tf
from tensorflow import keras
import time
import numpy as np
import zipfile
import python_pachyderm
from tensorflow.python.lib.io import file_io
from tensorflow.python.keras.utils.data_utils import get_file

# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir = os.getenv('DATA_DIR', "/tmp/data")
# this is the training data file in the input repo
training_data = os.getenv('TRAINING_DATA', "mninst.npz")
# this is the name of model file in the output repo
model_file = os.getenv('MODEL_FILE', "my_model.h5")

client = python_pachyderm.PfsClient(host=os.getenv('PACHD_SERVICE_HOST'), port=os.getenv('PACHD_SERVICE_PORT'))

def main(_):
  input_url = "s3://{}/".format(input_bucket)
  output_url = "s3://{}/".format(output_bucket)
  
  os.makedirs(data_dir)

  # first, we copy files from pachyderm into a convenient
  # local directory for processing.
  input_uri = os.path.join(input_url, training_data)
  training_data_path = os.path.join(data_dir, training_data)
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

    model.compile(
      optimizer=tf.keras.optimizers.Adam(),
      loss=tf.keras.losses.sparse_categorical_crossentropy,
      metrics=['accuracy']
    )

    return model

  # Create a basic model instance
  model = create_model()
  model.summary()

  model.fit(train_images, train_labels, batch_size=32, epochs=5, validation_data=(test_images, test_labels))

  # Save entire model to a HDF5 file
  model_file = os.path.join(data_dir, model_file)
  model.save(model_file)
  # Copy file over to Pachyderm
  output_uri = os.path.join(output_url, model_file)
  print("copying {} to {}".format(model_file, output_uri))
  file_io.copy(model_file, output_uri, True)

  client.finish_commit(('output', 'master'))

if __name__ == '__main__':
  tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
