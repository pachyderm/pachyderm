# Deploy the Pachyderm IDE

!!! note
    The Pachyderm IDE is an enterprise feature,
    which is also available for testing during
    the 14-day free trial.
    Contact sales@pachyderm.com for more
    information.

The Pachyderm Integrated Development Environment (IDE) is
an optional extension to Pachyderm clusters that provides a
comprehensive environment for prototyping and deploying
Pachyderm pipelines, as well as introspecting on data in the
Pachyderm version control system. It combines together Jupyter,
JupyterHub, and JupyterLab, which are familiar tools for many data
scientists.

JupyterLab is the newest Jupyter web interface that
allows you to run your Jupyter notebooks in tabs and extend the
tabs with custom applications. Our version of JupyterLab comes
with `pachctl`, `python-pachyderm`, Pachyderm shell,
Terminal, Python 3 Notebooks, and other useful components preconfigured.
From the terminal window, you can use `pachctl` commands and from 
within the notebooks you can call 
[Pachyderm Python client library](https://github.com/pachyderm/python-pachyderm)
methods to manage Pachyderm directly from the
JupyterLab UI.

The following diagram shows the Pachyderm and JupyterHub deployment.

![JupyterHub and Pachyderm Architecture Overview](../../assets/images/d_jupyterhub-pachyderm-arch.svg)

In the diagram above, you can see that Pachyderm and JupyterHub are
deployed on the same Kubernetes cluster. You deploy Pachyderm by
using the pachctl `deploy command` and JupyterHub by running
our deployment script as described below. After deployment, you log in
to JupyterHub with your Pachyderm user and interact with Pachyderm
from within the Jupyter UI by using the Pachyderm Python client.

## Prerequisites

Before deploying Pachyderm IDE, configure the following prerequisites:

* Deploy Pachyderm 1.11.0 or later as described in
[Deploy Pachyderm](../../deploy-manage/deploy/)
on a supported Kubernetes platforms:

  - Google Kubernetes Engine (GKE) with Kubernetes v1.13
  - Amazon Elastic Container Service (EKS) with Kubernetes v1.13
  - Docker Desktop for Mac

    For more information about JupyterHub requirements for Kubernetes,
    see [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/).

* Register your Enterprise token as described in
[Activate Pachyderm Enterprise Edition](../..//enterprise/deployment/#activate-pachyderm-enterprise-edition).

* Enable [Pachyderm Access Controls](../../enterprise/auth/auth/).

## Deploy Pachyderm IDE

After you deploy Pachyderm and enable authentication,
deploy the Pachyderm IDE by running:

```bash
pachctl deploy ide
```

Pachyderm deploys and configures JupyterHub and JupyterLab, which
might take some time. 

## Log in to Pachyderm IDE

After you deploy the Pachyderm IDE, you can access the JupyterLab UI
in a web browser through the Pachyderm cluster hostname on port
`80`. To get the IP address of the Pachyderm IDE,
run the following command:

* If you have deployed the Pachyderm IDE in a cloud platform, run:

  ```bash
  kubectl --namespace=default get svc proxy-public
  ```

* If you have deployed the Pachyderm IDE in a minikube, run:

  ```bash
  minikube service proxy-public --url
  ```

Paste the returned address in a browser to access your Pachyderm IDE.
Use your Pachyderm authentication token to log in.

If you access your Kubernetes cluster through a firewall, verify that
you can access your cluster on port 80. For more information, see
the documentation for your cloud platform in
[Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/create-k8s-cluster.html).


<--- TBA Manage JL instances through the JL UI-->

!!! note "See Also:"
    - [Use Pachyderm IDE](../../how-tos/use-pachyderm-ide/index.md)
