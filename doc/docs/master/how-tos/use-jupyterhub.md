# Use JupyterHub with `python-pachyderm`

This section describes how you can use the `python-pachyderm`
from within JupyterHub.

!!! note
    You need to have Pachyderm and JupyterHub instlalled on the
    same Kubernetes cluster as described in
    [Deploy Pachyderm with JupyterHub](deploy-pachyderm-jupyterhub.md).
    Currently, an existing JupyterHub deployment cannot be
    integrated with Pachyderm.

## Overview

JupyterHub is a popular data science platform that enables users
to quickly spin out multiple single-tenant Jupyter Notebook instances.
Jupyter Notebook provides a Python interactive development environment (IDE)
that is convenient for data science projects. Because of the built-in
rich-text support, visualizations, the easy-to-use web interface, many
enterprise users prefer Jupyter Notebooks to a classic Terminal prompt.
JupyterHub brings all the benefits of Jupyter Notebooks without the need
to install anything on their machines except for a web browser.

[python-pachyderm](https://github.com/pachyderm/python-pachyderm) is an
official Python client for Pachyderm.
The [API Documentation](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html)
describes various API operations that you can execute to interact with
Pachyderm.

When you deploy JupyterHub by using our deployment script, `python-pachyderm`
was enabled in JupyterHub so that you can run API requests directly from
your Jupyter Notebook.

The following diagram describes the Pachyderm and Jupyter Hub installation.

![JupyterHub and Pachyderm Architecture]()


<!-- TBA: Auth-->

### Difference in Pipeline Creation Methods

One of the biggest advantages of using `python-pachyderm` in conjunction
with JupyterHub is that you can avoid rebuilding the Docker image
every time you update your code. Because JupyterHub is a Python environment,
you can use the `create_python_pipeline` function to create pipeline from
within your JupyterHub Notebook.

When you use `pachctl` or the Pachyderm UI to create a pipeline, you execute
the [create_pipeline](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#python_pachyderm.Client.create_pipeline) method. Any time you need to update the code of your pipeline,
you need to build a new Docker image and push it to the a Docker registry.

When you use JupyterHub, in addition to the mentioned `create_pipelie` method,
you can use the [create_python_pipeline](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#python_pachyderm.create_python_pipeline) function that does not require
you to build a new image. Instead, it creates a PFS repository called
`pipeline_name.source` and puts the source code into it. Also, the function
creates a `pipeline_name.build` repository to store build Python dependecies.
Therefore, when you use `create_python_pipeline`, your DAG includes two
additional repositories for each pipeline. Because of that, you do not need
build a Docker image every time you change something your pipeline code.
You can run your code instantly.

While you can mix and match pipeline creation methods in JupyterHub, you might
eventually pick one method that works for your use case. It is a matter of
personal preference which method to use. While some users might find
convenient to avoid Docker build workflow, others might want to enable
Docker in JupyterHub. Alterntively, Docker images can be built and pushed
from your local machine.

In the [OpenCV example](), both methods are used in the same Jupyter
Notebook.

## Using `python-pachyderm` in JupyterHub

The `python-pachyderm` client is preinstalled in your JupyterHub instance
deployed as described in [Deploy Pachyderm with JupyterHub](deploy-pachyderm-jupyterhub.md).
This section describes basic operations that you can execute from JupyterHub
to interact with Pachyderm.

After you log in, use the [python-achyderm](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#header-functions)
client API to manage Pachyderm directly from your Jupyter notebooks.
The following code initializes the Python Pachyderm client in JupyterHub:

```bash
import python_pachyderm
client = python_pachyderm.Client.new_in_cluster()
```

!!! note
    This function is different from the function you'd call locally.

Then, you can use API requests to create Pachyderm repositories, pipelines,
put files, and others. To execute your code, select the corresponding cell
and click **Run**.

For example, you can check the current user by
running the following code:

```bash
import python_pachyderm
client = python_pachyderm.Client.new_in_cluster()
print(client.who_am_i())
```

The following screenshot demonstrates how this looks in JupyterHub:
![JupyterHub whoami](../../assets/images/s_jupyterhub_whoami.png)

### Create a Repostiory

To create a repository, run the following code:

```bash
import python_pachyderm
client = python_pachyderm.Client.new_in_cluster()
client.create_repo('test')
client.list_repo()
```

**System Response:**

```bash
[repo {
name: "test"
}
created {
  seconds: 1576869000
  nanos: 886123695
}
auth_info {
  access_level: OWNER
}
]
```

### Delete a Repository

To delete a repository, run the following code:

```bash
import python_pachyderm
client = python_pachyderm.Client.new_in_cluster()
client.delete_repo('test')
client.list_repo()
```

**System Response:**
```bash
[]
```

## Update Your Pipeline

When you need to update your pipeline,
you can do so directly in the JupyterHub UI by
modifying the corresponding notebook and
reruning it again.
If you use the `create_python_pipeline`
function that uses the code stored in a local directory,
you can update the pipeline directly in the JupyterHub
UI by adding the `update=True`
parameter into a new Jupyter notebook cell.

**Example:**

```bash hl_lines="6 10"
import os
import python_pachyderm
client = python_pachyderm.Client.new_in_cluster()
python_pachyderm.create_python_pipeline(
    client,
    "./edges",
    python_pachyderm.Input(pfs=python_pachyderm.PFSInput(glob="/*", repo="images")),
    update=True
)
```

If you are using the standard `create_pipeline` method, then
you need to rebuild and push your Docker container to your image
registry. Then, you need to update the 

