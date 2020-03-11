#!/usr/bin/python3

import kfp.dsl as kfp

@kfp.pipeline(
  name="mnist pipeline",
  description="Train neural net on MNIST"
)
def mnist_pipeline(s3_endpoint='',
                   input_bucket='input'):
  """
  Train neural net on 'mnist'
  """
  train = kfp.ContainerOp(
      name='mnist',
      image='pachyderm/mnist_klflow_example@sha256_',
      arguments=[
          "python3" "/app/tfjob.py",
          "--s3_endpoint", s3_endpoint,
          "--input_bucket", input_bucket,
      ]
  )

if __name__ == '__main__':
  import kfp.compiler as compiler
  compiler.Compiler().compile(mnist_pipeline, __file__ + '.tar.gz')
