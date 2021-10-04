# Local Installation

This guide covers how you can quickly get started using Pachyderm locally
on macOS®, Linux®, or Microsoft® Windows®. 

To install Pachyderm on Windows, take a look at
[Deploy Pachyderm on Windows](wsl-deploy.md) first.

!!! Warning
      - A local installation is **not designed to be a production
      environment**. It is meant to help you learn and experiment quickly with Pachyderm. 
      - A local installation is designed for a **single-node cluster**.
      This cluster uses local storage on disk and does not create a
      PersistentVolume (PV). If you want to deploy a production multi-node
      cluster, follow the instructions for your cloud provider or on-prem
      installation as described in [Deploy Pachyderm](../../deploy-manage/deploy/).
      New Kubernetes nodes cannot be added to this single-node cluster.
      - Pachyderm supports the **Docker runtime only**. If you want to
      deploy Pachyderm on a system that uses another container runtime,
      ask for advice in our [Slack channel](http://slack.pachyderm.io/).


Pachyderm uses `Helm` for all deployments.
## Prerequisites

The following prerequisites are required for a successful local deployment of Pachyderm:

- A Kubernetes cluster running on your local environment: 
      - [Docker Desktop](#using-kubernetes-on-docker-desktop),
      - [Minikube](#using-minikube)
      - [Kind](#using-kind)
      - Oracle® VirtualBox™
- [Helm](#install-helm)
- [Pachyderm Command Line Interface (`pachctl`)](#install-pachctl)
### Using Minikube

On your local machine, you can run Pachyderm in a minikube virtual machine.
Minikube is a tool that creates a single-node Kubernetes cluster. This limited
installation is sufficient to try basic Pachyderm functionality and complete
the Beginner Tutorial.

To configure Minikube, follow these steps:

1. Install minikube and VirtualBox in your operating system as described in
the [Kubernetes documentation](http://kubernetes.io/docs/getting-started-guides/minikube).
1. [Install `kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
1. Start `minikube`:

      ```shell
      minikube start
      ```

!!! Note
    Any time you want to stop and restart Pachyderm, run `minikube delete`
    and `minikube start`. Minikube is not meant to be a production environment
    and does not handle being restarted well without a full wipe.

### Using Kubernetes on Docker Desktop

If you are using Minikube, skip this section.

You can use Kubernetes on Docker Desktop instead of Minikube on macOS or Linux
by following these steps:

1. In the Docker Desktop Preferences, enable Kubernetes:
   ![Docker Desktop Enable K8s](../images/k8s_docker_desktop.png)

2. From the command prompt, confirm that Kubernetes is running:
   ```shell
   kubectl get all
   ```
   ```
   NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
   service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   5d
   ```

   * To reset your Kubernetes cluster that runs on Docker Desktop, click
   the **Reset Kubernetes cluster** button. See image above. 

### Using Kind

!!! Note
      Please note that Kind is *experimental* still.

1. Install Kind according to its [documentation](https://kind.sigs.k8s.io/).

1. From the command prompt, confirm that Kubernetes is running:
   ```shell
   kubectl get all
   ```
   ```
   NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
   service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   5d
   ```

### Install `pachctl`

`pachctl` is a command-line tool that you can use to interact
with a Pachyderm cluster in your terminal.

1. Run the corresponding steps for your operating system:

      * For macOS, run:

      ```shell
      brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}
      ```

      * For a Debian-based Linux 64-bit or Windows 10 or later running on
      WSL:

      ```shell
      curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
      ```

      * For all other Linux flavors:

      ```shell
      curl -o /tmp/pachctl.tar.gz -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_linux_amd64.tar.gz && tar -xvf /tmp/pachctl.tar.gz -C /tmp && sudo cp /tmp/pachctl_{{ config.pach_latest_version }}_linux_amd64/pachctl /usr/local/bin
      ```

1. Verify that installation was successful by running `pachctl version --client-only`:

      ```shell
      pachctl version --client-only
      ```

      **System Response:**

      ```shell
      COMPONENT           VERSION
      pachctl             {{ config.pach_latest_version }}
      ```

      If you run `pachctl version` without the flag `--client-only`, the command times
      out. This is expected behavior because Pachyderm has not been deployed yet (`pachd` is not yet running).

!!! Note "Architecture"
      A look at [Pachyderm high-level architecture diagram](https://docs.pachyderm.com/latest/deploy-manage/#overview) 
      will help you build a mental image of Pachyderm various architectural components.

### Install `Helm`

Follow Helm's [installation guide](https://helm.sh/docs/intro/install/).

## Deploy Pachyderm's latest version with Helm

When done with the [Prerequisites](#prerequisites),
deploy Pachyderm on your local cluster by following these steps:

!!! Tip
    If you are new to Pachyderm, try [Pachyderm Shell](../../deploy-manage/manage/pachctl_shell/).
    This add-on tool suggests `pachctl` commands as you type. 
    It will help you learn Pachyderm's main commands faster.

* Get the Repo Info:
   ```shell
   $ helm repo add pach https://helm.pachyderm.com
   ```
   ```shell
   $ helm repo update
   ```

* Install Pachyderm's latest helm chart ([helm v3](https://helm.sh/docs/intro/)):
   ```shell
   $ helm install pachd pach/pachyderm --set deployTarget=LOCAL
   ```

!!! Info "See Also"
      More [details on Pachyderm's Helm installation](../../deploy-manage/deploy/helm_install/).

## Check your install

Check the status of the Pachyderm pods by periodically
running `kubectl get pods`. When Pachyderm is ready for use,
all Pachyderm pods must be in the **Running** status.

Because Pachyderm needs to pull the Pachyderm Docker image
from DockerHub, it might take a few minutes for the Pachyderm pods status
to change to `Running`.


```shell
kubectl get pods
```

**System Response:**

```shell
NAME                                    READY   STATUS    RESTARTS   AGE
console-7f4b749444-78kzz                1/1     Running   0          6h
etcd-0                                  1/1     Running   0          6h
loki-0                                  1/1     Running   0          6h
loki-promtail-zz8ch                     1/1     Running   0          6h
pachd-5f6c956647-cj9g8                  1/1     Running   4          6h
postgres-0                              1/1     Running   0          6h
release-name-traefik-5659968869-v58j9   1/1     Running   0          6h
```

If you see a few restarts on the `pachd` nodes, that means that
Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
Kubernetes restarted those pods. Re-run `kubectl get pods`

## Have 'pachctl' and your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

* If you exposed your cluster to the internet by setting up a LoadBalancer in the `values.yaml` as follow:

     ```yaml
     pachd:
      service:
        type: LoadBalancer
     ```

    1. Retrieve the external IP address of the service.  When listing your services again, you should see an external IP address allocated to the `pachd` service 

        ```shell
        $ kubectl get service
        ```

    1. Update the context of your cluster with their direct url, using the external IP address above:

        ```shell
        $ echo '{"pachd_address": "grpc://<external-IP-address>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
        ```

    1. Check that your are using the right context: 

        ```shell
        $ pachctl config get active-context`
        ```

        Your cluster context name should show up.


* If you're not exposing `pachd` publicly, you can run:

    ```shell
    # Background this process because it blocks.
    $ pachctl port-forward
    ``` 
    Open a new terminal window. This command does not exit unless you interrupt it.

* Verify that `pachctl` and your cluster are connected:

    ```shell
    $ pachctl version
    ```

    **System Response:**

    ```
    COMPONENT           VERSION
    pachctl             {{ config.pach_latest_version }}
    pachd               {{ config.pach_latest_version }}
    ```

## Next Steps

* Complete the [Beginner Tutorial](./beginner_tutorial.md)
to learn the basics of Pachyderm, such as adding data and building
analysis pipelines.

* Explore the Pachyderm Console.
By default, Pachyderm deploys the Pachyderm Enterprise Console. You can
use a FREE trial token to experiment with it. Point your
browser to port `30080` on your minikube IP.
Alternatively, if you cannot connect directly, enable port forwarding
by running `pachctl port-forward`, and then point your browser to
`localhost:30080`.

!!! note "See Also:"
    [General Troubleshooting](../troubleshooting/general_troubleshooting.md)
