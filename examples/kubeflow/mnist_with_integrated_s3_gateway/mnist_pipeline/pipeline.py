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
      image='pachyderm/mnist_klflow_example@sha256:824651c1ba03faa1d7d0ec2899d77fad7da4f84c4ad60aa6df8fee042f6cb7e2',
      arguments=[
          "python3",
          "/app/tfjob.py",
          "--s3_endpoint", s3_endpoint,
          "--input_bucket", "input",
      ]
  )

if __name__ == '__main__':
  import kfp.compiler as compiler
  compiler.Compiler().compile(mnist_pipeline, __file__ + '.tar.gz')
