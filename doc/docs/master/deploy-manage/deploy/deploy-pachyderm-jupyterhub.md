# Deploy Pachyderm IDE

!!! note
    Pachyderm IDE is an enterprise feature.
    Contact sales@pachyderm.com for more
    information.

Pachyderm integrated development environment (IDE) is
an extension to the standard Pachyderm deployment that
enables you to leverage JupyterLab user interface
seamlessly integrated with Pachyderm data version control.
The data scientists that have been using Jupyter products in
the past can now enjoyh this new and smooth development
experience.

When you deploy the Pachyderm IDE, you spin up a JupyterHub
deployment in the same Kubernetes cluster where Pachyderm
runs. JupyterHub is an open-source platform that enables you
to create multiple instances of single-user Jupyter notebook
servers on-demand for each member of your team.
This way, each user gets a personal notebook server.

JupyterLab is the newest Jupyter notebook web interface that
allows you to run your Jupyter notebooks in tabs and configure
tabs as needed. Our version of JupyterLab comes with Pachyderm shell,
Terminal, Python 3 Notebooks, and other useful tabs preconfigured.
From the terminal window you can use `pachctl` commands and from 
within the notebooks you can call [Pachyderm Python client library](https://github.com/pachyderm/python-pachyderm) methods to manage Pachyderm directly from the
JupyterLab UI.

By using Pachyderm authentication, you log in to JupyterHub with
your Pachyderm credentials. In essence, you run your code
in JupyterLab and use the `python-pachyderm` API to create your
pipelines and version your data in Pachyderm from within
the JupyterLab UI.

The following diagram shows the Pachyderm and JupyterHub deployment.

![JupyterHub and Pachyderm Architecture Overview](../../assets/images/d_jupyterhub-pachyderm-arch.svg)

In the diagram above, you can see that Pachyderm and JupyterHub are
deployed on the same Kubernetes cluster. You deploy Pachyderm by
using the pachctl `deploy command` and JupyterHub by running
our deployment script as described below. After deployment, you log in
to JupyterHub with your Pachyderm user and interact with Pachyderm
from within the Jupyter UI by using the Pachyderm Python client.

## Prerequisites

Before deploying Pachyderm IDE, you need to make sure that you configure
the following prerequisites:

* Deploy Pachyderm 1.11.0 or later as described in [Deploy Pachyderm](https://docs.pachyderm.com/latest/deploy-manage/deploy/) on a supported Kubernetes platforms:

  - Google Kubernetes Engine (GKE) with Kubernetes v1.13
  - Amazon Elastic Container Service (EKS) with Kubernetes v1.13
  - Docker Desktop for Mac with Kubernetes v1.14

    For more information about JupyterHub requirements for Kubernetes,
    see [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/).

* Register your Enterprise token as described in
[Activate Pachyderm Enterprise Edition](https://docs.pachyderm.com/latest/enterprise/deployment/#activate-pachyderm-enterprise-edition).
* Enable [Pachyderm Access Controls](https://docs.pachyderm.com/latest/enterprise/auth/auth/).

### Enable Pachyderm Access Controls (Optional) - is this now required?

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


## Deploy Pachyderm IDE

After you deploy Pachyderm and enable authentication,
deploy Pachyderm IDE by running:

```bash
pachctl deploy ide
```

The installation might take some time. 

## Log in to Pachyderm IDE

After you deploy JupyterHub, you can access the JupyterHub UI
in a web browser through the Pachyderm cluster hostname on port
`80`. To get your the IP address of your JupyterHub deployment,
run the following command:

* If you have deployed Pachyderm IDE in a cloud platform, run:

  ```bash
  kubectl --namespace=default get svc proxy-public
  ```

* If you have deployed Pachyderm IDE in a minikube, run:

  ```bash
  minikube service proxy-public --url
  ```

Paste the returned address in a browser to access your Pachyderm IDE.
Use your Pachyderm token to log in.

If you access your Kubernetes cluster through a firewall, verify that
you can access your cluster on port 80. For more information, see
the documentation for your cloud platform in
[Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/create-k8s-cluster.html).

!!! note "See Also:"
    - [Use JupyterHub](../../how-tos/use-jupyterhub/index.md)
