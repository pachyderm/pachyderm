# Deploy Pachyderm with JupyterHub

JupyterHub is an open-source platform that enables you
spin multiple instances of single-user Jupyter notebook
servers on-demand for each member of your team.
This way, each user gets their own isolated notebook server
to work on.

You can deploy JupyterHub alongside Pachyderm
on the same Kubernetes cluster and use Pachyderm authentication
and Pachyderm [Python client library]() with JupyterHub
By using Pachyderm authentication, you can log in to JupyterHub with
your Pachyderm credentials. In essence, you store your code and
data in JupyterHub before you put it in Pachyderm for versioning.

## Deploy Prerequisites

Before you can deploy JupyterHub with Pachyderm, you need to
install Pachyderm on a supported Kubernetes platform and Helm 3
on your client machine.

### Deploy Pachyderm

You need to deploy Pachyderm as described in [Deploy Pachyderm]()
in of the supported platforms:

- Google Kubernetes Engine (GKE) with Kubernetes v1.13
- Amazon Elastic Container Service (EKS) with Kubernetes v1.13
- Docker Desktop for Mac with Kubernetes v1.14

!!! note
    Kubernetes v1.16 is not supported including Minikube deployments.
    If you already have a local Pachyderm deployed with Kubernetes v1.16,
    you might not be able to deploy JupyterHub on the same Kubernetes
    cluster by using our deployment script.

For more information about JupyterHub requirements for Kubernetes,
see [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/).

If you access your Kubernetes cluster through a firewall, verify that
you configure networking parameters so that you can access your cluster
on port 80. For more information, see the documentation for your
cloud in [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/create-k8s-cluster.html).

### Enable Pachyderm Access Controls (Optional)

You can enable Pachyderm access controls to use Pachyderm credentials
to log in to JupyterHub. However, this
is not required, and if you prefer to use JupyterHub authentication
instead, skip this section.

Pachyderm Access Controls is an enterprise feature. You need to have
an active Enterprise token to enable it.

To enable Pachyderm authentication, follow the steps in
[Configure Access Controls](https://docs.pachyderm.com/latest/enterprise/auth/auth/)

!!! note
    If you decide to enable Pachyderm authentication later, you
    will need to either manually reconfigure or redeploy JupyterHub.

### Install Helm

Helm is a package manager for Kubernetes that controls JupyterHub
installation.

To install Helm on your client machine, follow the steps in the
[Helm Documentation](https://helm.sh/docs/intro/install/).

## Deploy JupyterHub

After you deploy Pachyderm in one of the supported environments and
install Helm on your client machine, you can use the deployment script
to deploy JupyterHub.

To deploy JupyterHub, complete the following steps:

1. Clone the [jypethub-pachyderm]( https://github.com/pachyderm/jupyterhub-pachyderm)
repository:

   ```bash
   git clone git@github.com:pachyderm/jupyterhub-pachyderm.git
   ```

1. Change the current directory to `jupyterhub-pachyderm`:

   ```bash
   cd jupyterhub-pachyderm
   ```

1. Run the `./init.py` script:

   ```bash
   ./init.py
   ```

   **System response:**

   ```bash
   ===> checking dependencies are installed
   ...
   ===> installing jupyterhub
   Release "jhub" does not exist. Installing it now.
   NAME: jhub
   LAST DEPLOYED: Wed Dec 18 13:09:26 2019
   NAMESPACE: default
   STATUS: deployed
   REVISION: 1
   TEST SUITE: None
   NOTES:
   Thank you for installing JupyterHub!
   ...
   ===> notes
   - Since Pachyderm auth doesn't appear to be enabled, JupyterHub will expect the following global password for users: XXXXXXXXXXXXXXXXXXXXX
   ```

   The JupyterHub is now deployed on your Kubernetes cluster.

## Log in to JupyterHub

After you deploy JupyterHub, you can access the JupyterHub UI
in a web browser through the Pachyderm cluster hostname on port
`80`.

If you have enabled Pachyderm access controls, use your Pachyderm
token to log in. Otherwise, use the global password
that was printed out when you ran the `./init.py` script with any
username.

## Using `python-pachyderm`

After you log in you can use [Python Pachyderm](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#header-functions)
client API to manage Pachyderm directly in Jupyter notebooks.

For example, to create a repository in Pachyderm, complete the following
steps:

1. In the JupyterHub UI, create a new notebook by clicking **New**
and selecting **Python 3**.
1. In the dialog field, type:

   ```bash

   ```


