# Use the Pachyderm IDE with `python-pachyderm`

!!! note
    The Pachyderm IDE is an enterprise feature,
    which is also available for testing during
    the 14-day free trial.
    Contact sales@pachyderm.com for more information.

This section describes how you can use the `python-pachyderm`
client from within the Pachyderm IDE.

!!! note
    You need to have Pachyderm IDE installed as described in
    [Deploy the Pachyderm IDE](../../deploy-manage/deploy/deploy-pachyderm-ide/).

## Overview

When you deploy the Pachyderm IDE, you get JupyterHub and
a customized JupyterLab UI running next to your Pachyderm
cluster.

Before we proceed, let's clarify the following terms
as they might be confusing for first-time users:

* JupyterHub is a popular data science platform that enables users
to quickly spin out multiple single-tenant Jupyter Notebook server instances.

* Jupyter Notebooks provide a means for conducting experiments with data and
code written in Python which is familiar to many data scientists. Because of
the built-in rich-text support, visualizations, the easy-to-use web interface,
many enterprise users prefer Jupyter Notebooks to the classic Terminal prompt.
JupyterHub brings all the benefits of Jupyter Notebooks without the need
to install or configure anything on user machines except for a web browser.

* JupyterLab is an alternative UI for the classic Jupyter Notebook IDE
that enables you to author and test notebooks and
code.

* [python-pachyderm](../../reference/clients/#python-client) is an
official Python client for Pachyderm. For Python developers who prefer to
communicate with Pachyderm directly through the API, rather than by using
the `pachctl` tool, `python-pachyderm` is the right choice.

`python-pachyderm` is preinstalled in your Pachyderm IDE.

### Difference in Pipeline Creation Methods

`python-pachyderm` supports the standard
[create_pipeline](https://python-pachyderm.readthedocs.io/en/v6.x/python_pachyderm.mixin.html?highlight=create_pipeline#python_pachyderm.mixin.pps.PPSMixin.create_pipeline)
method that is
also available through the Pachyderm CLI and UI. When you use
`create_pipeline`, you need to build a new Docker image and push
it to an image registry every
time you update the code in your pipeline. Users that are less familiar
with Docker might find this process a bit cumbersome. However, you must
use this method for all non-Python code.

When you use `python-pachyderm`, in addition to the
`create_pipeline` method,
you can use the [create_python_pipeline](https://python-pachyderm.readthedocs.io/en/v6.x/python_pachyderm.html#python_pachyderm.util.create_python_pipeline)
function that does not require
you to include your code in a Docker image and rebuild it each time you make
a change. Instead, this function creates a PFS repository
called `<pipeline_name>_source` and puts the source code into it. Also, it
creates a `<pipeline_name>_build` repository to build Python dependencies.
Therefore, when you use `create_python_pipeline`, your DAG includes two
additional repositories for each pipeline.
Because of that, you do not need
to build a new Docker image every time you change something in your
pipeline code. You can run your code instantly. This method is ideal for
users who want to avoid building Docker images.

While you can mix and match pipeline creation methods in Pachyderm IDE, you
might eventually want to pick one method that works for your use case. It is a
matter of personal preference which method to use. While some users, those
who write code in Python, in particular, might
find it convenient to avoid the Docker build workflow, others might want to
enable Docker in JupyterHub or build Docker images from their local machines.


