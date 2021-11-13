# Deploy Pachyderm with JupyterHub

!!! note
    JupyterHub integration with Pachyderm is an
    enterprise feature. Contact sales@pachyderm.com
    to request enabling JupyterHub integration
    for your Pachyderm Enterprise license.

JupyterHub is an open-source platform that enables you
to spin multiple instances of single-user Jupyter notebook
servers on-demand for each member of your team.
This way, each user gets a personal notebook server.

You can deploy JupyterHub alongside Pachyderm
on the same Kubernetes cluster and use Pachyderm authentication
and [Pachyderm Python client library](https://github.com/pachyderm/python-pachyderm)
with JupyterHub.
By using Pachyderm authentication, you can log in to JupyterHub with
your Pachyderm credentials. In essence, you run your code
in JupyterHub and use the `python-pachyderm` API to create your
pipelines and version your data in Pachyderm from within
the JupyterHub UI.

The following diagram shows the Pachyderm and JupyterHub deployment.

![JupyterHub and Pachyderm Architecture Overview](../../assets/images/d_jupyterhub-pachyderm-arch.svg)

In the diagram above, you can see that Pachyderm and JupyterHub are
deployed on the same Kubernetes cluster. You deploy Pachyderm by
using the pachctl `deploy command` and JupyterHub by running
our deployment script as described below. After deployment, you log in
to JupyterHub with your Pachyderm user and interact with Pachyderm
from within the Jupyter UI by using Pachyderm Python client.

## Prerequisites

Before you can deploy JupyterHub with `python-pachyderm`, you need to
install Pachyderm on a supported Kubernetes platform and configure
other prerequisites listed below.

### Deploy Pachyderm

You need to deploy Pachyderm as described in
[Deploy Pachyderm](../../deploy/)
in of the supported platforms:

- Google Kubernetes Engine (GKE) with Kubernetes v1.13
- Amazon Elastic Container Service (EKS) with Kubernetes v1.13
- Docker Desktop for Mac with Kubernetes v1.14

!!! note
    Kubernetes v1.16 is not supported, including Minikube deployments.
    If you already have a local Pachyderm deployed with Kubernetes v1.16,
    you might not be able to deploy JupyterHub on the same Kubernetes
    cluster by using the Jupyter deployment script.

For more information about JupyterHub requirements for Kubernetes,
see [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/).

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
    Your existing notebooks will be preserved after
    the redeploy.

### Install Helm

Helm is a package manager for Kubernetes that controls the JupyterHub
installation. Our deployment script requires Helm 3 or later.

To install Helm on your client machine, follow the steps in the
[Helm Documentation](https://helm.sh/docs/intro/install/).

## Deploy JupyterHub

After you deploy Pachyderm in one of the supported environments and
install Helm on your client machine, you can use the deployment script
to deploy JupyterHub.

To deploy JupyterHub, complete the following steps:

<!--1. Clone the [jupyterhub-pachyderm](https://github.com/pachyderm/jupyterhub-pachyderm)
repository:

   ```shell
   git clone git@github.com:pachyderm/jupyterhub-pachyderm.git
   ```

1. Change the current directory to `jupyterhub-pachyderm/`:

   ```shell
   cd jupyterhub-pachyderm
   ```
-->
1. Run the `./init.py` script:

   ```shell
   ./init.py
   ```

   **System response:**

   ```
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

   JupyterHub is now deployed on your Kubernetes cluster.

## Log in to JupyterHub

After you deploy JupyterHub, you can access the JupyterHub UI
in a web browser through the Pachyderm cluster hostname on port
`80`.

If you have enabled Pachyderm access controls, use your Pachyderm
token to log in. Otherwise, use the global password
that was printed out in the terminal window when you ran the
`./init.py` script. You can use any arbitrary username.

If you access your Kubernetes cluster through a firewall, verify that
you can access your cluster on port 80. For more information, see
the documentation for your cloud platform in
[Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/create-k8s-cluster.html).

!!! note "See Also:"
    - [Use JupyterHub](../../how-tos/use-jupyterhub/index.md)
