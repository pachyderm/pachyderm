#!/usr/bin/python3

import kfp.dsl as kfp
from kubernetes.client.models import V1EnvVar

@kfp.pipeline(
    name="mnist kubeflow pipeline",
    description="Train neural net on MNIST"
)
def kubeflow_pipeline(s3_endpoint: str = '', input_bucket : str = 'input'):
    """
    Train neural net on 'mnist'
    """
    pipeline = kfp.ContainerOp(
        name='mnist',
        image='ysimonson/mnist_kubeflow_pipeline:v1.0.0',
        arguments=["python3", "/app/tfjob.py"]
    )
    pipeline = pipeline.add_env_variable(V1EnvVar(name='S3_ENDPOINT', value=s3_endpoint))
    pipeline = pipeline.add_env_variable(V1EnvVar(name='INPUT_BUCKET', value=input_bucket))
    return pipeline

if __name__ == '__main__':
    import kfp.compiler as compiler
    compiler.Compiler().compile(kubeflow_pipeline, __file__ + '.tar.gz')
