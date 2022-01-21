# Deploy the Pachyderm IDE

!!! Note
    The Pachyderm IDE is an enterprise feature. Request an [**Enterprise Edition trial token**](https://www.pachyderm.com/trial/).

The Pachyderm Integrated Development Environment (IDE) is
an optional extension to Pachyderm clusters that provides a
comprehensive environment for prototyping and deploying
Pachyderm pipelines, as well as introspecting on data in the
Pachyderm version control system. It combines together Jupyter,
JupyterHub, and JupyterLab, which are familiar tools for many data
scientists.

The following diagram shows the Pachyderm IDE deployment.

![JupyterHub and Pachyderm Architecture Overview](../../assets/images/d_jupyterhub-pachyderm-arch.svg)

In the diagram above, you can see that Pachyderm and JupyterHub are
deployed on the same Kubernetes cluster. You deploy Pachyderm by
using the `pachctl deploy` command as described below. After
deployment, you log in to JupyterHub with your Pachyderm user
and interact with Pachyderm from within the JupyterLab UI by
using `pachctl` or the Pachyderm Python client.

## Prerequisites

Before deploying Pachyderm IDE, configure the following prerequisites:

* Deploy Pachyderm {{ config.pach_latest_version }} or later as described in
[Deploy Pachyderm](../)
on a supported Kubernetes platforms:

    - Google Kubernetes Engine (GKE)
    - Amazon Elastic Kubernetes Service (EKS)
    - Docker Desktop for Mac
    - Minikube

    For more information about JupyterHub requirements for Kubernetes,
    see [Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/).

* Register your Enterprise token as described in
[Activate Pachyderm Enterprise Edition](../../../enterprise/deployment/#activate-pachyderm-enterprise-edition).

* Enable [Pachyderm Access Controls](../../../enterprise/auth/enable-auth/).

## Deploy Pachyderm IDE

After you deploy Pachyderm and enable authentication,
deploy the Pachyderm IDE by running:

```shell
pachctl deploy ide
```

Pachyderm deploys and configures JupyterHub and JupyterLab, which
might take some time. 

## Log in to Pachyderm IDE

After you deploy the Pachyderm IDE, you can access the UI
in a web browser through its service IP on port
`80`. To get the service IP address of the Pachyderm IDE,
run the following command:

* If you have deployed the Pachyderm IDE in a cloud platform, run:

  ```shell
  kubectl --namespace=default get svc proxy-public
  ```

* If you have deployed the Pachyderm IDE in Minikube, run:

  ```shell
  minikube service proxy-public --url
  ```

Paste the returned address in a browser to access your Pachyderm IDE.
Use your Pachyderm authentication token to log in.

If you access your Kubernetes cluster through a firewall, verify that
you can access your cluster on port 80. For more information, see
the documentation for your cloud platform in
[Zero to JupyterHub with Kubernetes](https://zero-to-jupyterhub.readthedocs.io/en/latest/create-k8s-cluster.html).


!!! note "See Also:"
    - [Use Pachyderm IDE](../../how-tos/use-pachyderm-ide/index.md)
