#!/usr/bin/python3

import json

import kfp.dsl as kfp
from kubernetes.client.models import V1EnvVar

@kfp.pipeline(
    name="mnist kubeflow pipeline",
    description="Train neural net on MNIST"
)
def kubeflow_pipeline(s3_endpoint, input_bucket):
    """
    Train neural net on 'mnist'
    """

    with open("../version.json") as f:
        j = json.load(f)
        pipeline = kfp.ContainerOp(
            name='mnist',
            image='ysimonson/mnist_kubeflow_pipeline:v{}'.format(j["kubeflow_pipeline"]),
            arguments=["python3", "/app/tfjob.py"]
        )

    pipeline = pipeline.add_env_variable(V1EnvVar(name='S3_ENDPOINT', value=s3_endpoint))
    pipeline = pipeline.add_env_variable(V1EnvVar(name='INPUT_BUCKET', value=input_bucket))

    return pipeline

if __name__ == '__main__':
    import kfp.compiler as compiler
    compiler.Compiler().compile(kubeflow_pipeline, __file__ + '.tar.gz')
