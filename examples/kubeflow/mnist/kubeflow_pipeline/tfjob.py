from __future__ import absolute_import, division, print_function, unicode_literals

import os
import sys
import tempfile

import boto3
import tensorflow as tf
from tensorflow import keras

def main(_):
    # TODO: remove
    for k, v in sorted(os.environ.items()):
        print("{}={}".format(k, v))

    s3_client = boto3.client(
        's3',
        endpoint_url=os.environ["S3_ENDPOINT"],
        aws_access_key_id='',
        aws_secret_access_key=''
    )

    input_bucket = os.getenv('INPUT_BUCKET', "input")

    with tempfile.TemporaryDirectory(suffix="pachyderm-mnist") as data_dir:
        # first, we copy files from pachyderm into a convenient
        # local directory for processing.
        training_data_path = os.path.join(data_dir, "mnist.npz")
        print("copying from {} to {}".format(input_bucket, training_data_path))
        s3_client.download_file(input_bucket, "mnist.npz", training_data_path)

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
        print("copying {} to {}".format(model_path, model_url))
        s3_client.put_object(model_path, "out", "my_model.h5")

if __name__ == '__main__':
    tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
