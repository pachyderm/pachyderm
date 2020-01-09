# Use JupyterHub with `python-pachyderm`

This section describes how you can use the `python-pachyderm`
client from within the JupyterHub UI.

!!! note
    You need to have Pachyderm and JupyterHub instlalled on the
    same Kubernetes cluster as described in
    [Deploy Pachyderm with JupyterHub](../../deploy-manage/deploy/deploy-pachyderm-jupyterhub.md).
    Currently, an existing JupyterHub deployment cannot be
    integrated with Pachyderm.

## Overview

JupyterHub is a popular data science platform that enables users
to quickly spin out multiple single-tenant Jupyter Notebook server instances.
Jupyter Notebook provides a Python interactive development environment (IDE)
that is convenient for data science projects. Because of the built-in
rich-text support, visualizations, the easy-to-use web interface, many
enterprise users prefer Jupyter Notebooks to the classic Terminal prompt.
JupyterHub brings all the benefits of Jupyter Notebooks without the need
to install or configure anything on user machines except for a web browser.

[python-pachyderm](https://github.com/pachyderm/python-pachyderm) is an
official Python client for Pachyderm. For Python developers who prefer to
communicate with Pachyderm directly through the API, rather than by using
the `pachctl` tool, `python-pachyderm` is the right choice.
The [API Documentation](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html)
describes various API operations that you can execute to interact with
Pachyderm.

When you deployed JupyterHub by using our deployment script, `python-pachyderm`
was enabled in JupyterHub so that you can run API requests directly from
your Jupyter Notebook.

### Difference in Pipeline Creation Methods

One of the most significant advantages of using `python-pachyderm`
with JupyterHub is that you can avoid rebuilding your Docker images
every time you update your code. Because JupyterHub is a Python environment,
you can use the `create_python_pipeline` function to create a pipeline from
within your JupyterHub Notebook and store your code locally rather than
in a Docker container.

When you use `pachctl` or the Pachyderm UI to create a pipeline, you execute
the
[create_pipeline](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#python_pachyderm.Client.create_pipeline)
method in the background. This standard workflow requires you to build a new
Docker image and push it to an image registry every time you update the
code in your pipeline. Users that are less familiar with Docker might find this
process a bit cumbersome.

When you use JupyterHub, in addition to the mentioned `create_pipeline` method,
you can use the [create_python_pipeline](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#python_pachyderm.create_python_pipeline) function that does not require
you to include your code in a Docker image and rebuild it each time you make
a change. Instead, this function creates a PFS repository
called `pipeline_name.source` and puts the source code into it. Also, it
creates a `pipeline_name.build` repository to build Python dependencies.
Therefore, when you use `create_python_pipeline`, your DAG includes two
additional repositories for each pipeline. While you still need a generic
image for your pipeline, your code is stored separately from it.
Because of that, you do not need
to build a new Docker image every time you change something in your
pipeline code. You can run your code instantly. This method is ideal for
users who want to avoid building Docker images.

While you can mix and match pipeline creation methods in JupyterHub, you might
eventually want to pick one method that works for your use case. It is a
matter of personal preference which method to use. While some users might find
convenient to avoid the Docker build workflow, others might want to enable
Docker in JupyterHub or build Docker images from their local machines.

In the [OpenCV Example for JupyterHub](https://github.com/pachyderm/jupyterhub-pachyderm),
both methods are used in the same notebook cell.
