from __future__ import absolute_import, division, print_function, unicode_literals

import os
import sys
import tempfile

import tensorflow as tf
from tensorflow import keras
from tensorflow.python.lib.io import file_io

def main(_):
    input_bucket = os.getenv('INPUT_BUCKET', "input")
    input_url = "s3://{}/".format(input_bucket)

    with tempfile.TemporaryDirectory(suffix="pachyderm-mnist-with-tfjob") as data_dir:
        # first, we copy files from pachyderm into a convenient
        # local directory for processing.
        training_data_url = os.path.join(input_url, "mnist.npz")
        training_data_path = os.path.join(data_dir, "mnist.npz")
        print("copying {} to {}".format(training_data_url, training_data_path))
        file_io.copy(training_data_url, training_data_path, True)

        (train_images, train_labels), (test_images, test_labels) = tf.keras.datasets.mnist.load_data(path=training_data_path)
        train_labels = train_labels[:1000]
        test_labels = test_labels[:1000]

        train_images = train_images[:1000].reshape(-1, 28 * 28) / 255.0
        test_images = test_images[:1000].reshape(-1, 28 * 28) / 255.0

        # Create a basic model instance
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

        model.summary()

        model.fit(train_images, train_labels, batch_size=32, epochs=5, validation_data=(test_images, test_labels))

        # Save entire model to a HDF5 file
        model_path = os.path.join(data_dir, "my_model.h5")
        model.save(model_path)
        # Copy file over to Pachyderm
        model_url = os.path.join("s3://out/", "my_model.h5")
        print("copying {} to {}".format(model_path, model_url))
        file_io.copy(model_path, model_url, True)

if __name__ == '__main__':
    tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
