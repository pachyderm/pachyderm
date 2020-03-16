#!/usr/bin/env python3

import os
import json

import kfp
import kfp.dsl
import kfp.components
from kubernetes.client.models import V1EnvVar

def mnist(s3_endpoint: str, input_bucket: str):
    # install boto3 - kinda nasty, but this is the way to do it in kubeflow's
    # lightweight components
    import sys, subprocess
    subprocess.run([sys.executable, '-m', 'pip', 'install', 'boto3'])

    # imports are done here because it's required for kubeflow's lightweight
    # components:
    # https://www.kubeflow.org/docs/pipelines/sdk/lightweight-python-components/
    import os
    import sys
    import tempfile

    import boto3
    import tensorflow as tf
    from tensorflow import keras

    s3_client = boto3.client(
        's3',
        endpoint_url=s3_endpoint,
        aws_access_key_id='',
        aws_secret_access_key=''
    )

    with tempfile.TemporaryDirectory(suffix="pachyderm-mnist") as data_dir:
        # first, we copy files from pachyderm into a convenient
        # local directory for processing.
        training_data_path = os.path.join(data_dir, "mnist.npz")
        print("copying from s3g to {}".format(training_data_path))
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
        print("copying {} to s3g".format(model_path))
        s3_client.upload_file(model_path, "out", "my_model.h5")

@kfp.dsl.pipeline(
    name="mnist kubeflow pipeline",
    description="Train neural net on MNIST"
)
def kubeflow_pipeline(s3_endpoint: str, input_bucket: str):
    op = kfp.components.func_to_container_op(
        mnist,
        base_image='tensorflow/tensorflow:1.14.0-py3'
    )

    pipeline = op(s3_endpoint, input_bucket) \
        .add_env_variable(V1EnvVar(name='S3_ENDPOINT', value=s3_endpoint)) \
        .add_env_variable(V1EnvVar(name='INPUT_BUCKET', value=input_bucket))
    
    return pipeline

def main():
    client = kfp.Client()

    run_id = client.create_run_from_pipeline_func(
        kubeflow_pipeline,
        arguments={
            "s3_endpoint": os.environ["S3_ENDPOINT"],
            "input_bucket": "input",
        }
    ).run_id

    print("run id: {}".format(run_id))

    j = client.wait_for_run_completion(run_id, 60)
    print("status: {}".format(j.run.status))
    assert j.run.status != 'Failed'

if __name__ == "__main__":
    main()
