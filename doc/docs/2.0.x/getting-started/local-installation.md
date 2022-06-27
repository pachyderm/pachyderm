
# Local Installation  
  
This guide covers how you can quickly get started using Pachyderm locally on macOS®, Linux®, or Microsoft® Windows®. To install Pachyderm on Windows, first look at [Deploy Pachyderm on Windows](../wsl-deploy){target=_blank}.

Pachyderm is an application written in go that runs on top of a Kubernetes cluster. 
A common way to interact with Pachyderm is by using Pachyderm command-line tool `pachctl`, from a terminal window. To check the state of your deployment, you will also need to install `kubectl`, Kubernetes command-line tool. 

Additionally, we will show you how to deploy Pachyderm Web UI **[Console](../../deploy-manage/deploy/console){target=_blank}** on your local cluster. **Console** will help you visualize your DAGs (Directed Acyclic Graphs), monitor your pipeline executions, access your logs, and troubleshoot while your pipelines are running.

!!! Warning  
      - A local installation is **not designed to be a production  
      environment**. It is meant to help you learn and experiment quickly with Pachyderm.   
      - A local installation is designed for a **single-node cluster**.  
      This cluster uses local storage on disk and does not create  
      Persistent Volumes (PVs). If you want to deploy a production multi-node  
      cluster, follow the instructions for your cloud provider or on-prem  
      installation as described in [Deploy Pachyderm](../../deploy-manage/deploy/).  
      New Kubernetes nodes cannot be added to this single-node cluster.   
  
  
Pachyderm uses `Helm` for all deployments.  
## Prerequisites  
  
For a successful local deployment of Pachyderm, you will need:  
  
- A [Kubernetes cluster](#setup-a-local-kubernetes-cluster) running on your local environment:   
      - [Docker Desktop](#using-kubernetes-on-docker-desktop),  
      - [Minikube](#using-minikube)  
      - [Kind](#using-kind)  
      - Oracle® VirtualBox™   
- [Helm](#install-helm) to deploy Pachyderm on your Kubernetes cluster.  
- [Pachyderm Command Line Interface (`pachctl`)](#install-pachctl) to interact with your Pachyderm cluster.
- [Kubernetes Command Line Interface `kubectl`](https://kubernetes.io/docs/tasks/tools/){target=_blank} to interact with your underlying Kubernetes cluster.

### Setup A Local Kubernetes Cluster

Pick the virtual machine of your choice.

#### Using Minikube  
  
On your local machine, you can run Pachyderm in a minikube virtual machine.  
Minikube is a tool that creates a single-node Kubernetes cluster. This limited  
installation is sufficient to try basic Pachyderm functionality and complete  
the Beginner Tutorial.  
  
To configure Minikube, follow these steps:  
  
1. Install minikube and VirtualBox in your operating system as described in  
the [Kubernetes documentation](https://kubernetes.io/docs/setup/){target=_blank}.    
1. Start `minikube`:  
  
      ```shell  
      minikube start  
      ```  
      Linux users, add this `--driver` flag:
      ```shell
      minikube start --driver=kvm2
      ```
!!! Note
    Any time you want to stop and restart Pachyderm, run `minikube delete`  
    and `minikube start`. Minikube is not meant to be a production environment  
    and does not handle being restarted well without a full wipe.  
  
#### Using Kubernetes on Docker Desktop   
  
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
  
#### Using Kind  
  
1. Install Kind according to its [documentation](https://kind.sigs.k8s.io/){target=_blank}.  
  
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

!!! Tip  
    If you are new to Pachyderm, try [Pachyderm Shell](../../deploy-manage/manage/pachctl-shell/){target=_blank}. This add-on tool suggests `pachctl` commands as you type. It will help you learn Pachyderm's main commands faster.  

!!! Note "Architecture"  
      A look at [Pachyderm high-level architecture diagram](../../deploy-manage/#overview)   
      will help you build a mental image of Pachyderm various architectural components.  
  
      For information, you can also check what a production setup looks like in this [infrastructure diagram](../../deploy-manage/deploy/ingress/#deliver-external-traffic-to-pachyderm).  
  
### Install `Helm`  
  
Follow Helm's [installation guide](https://helm.sh/docs/intro/install/){target=_blank}.  
  
## Deploy Pachyderm Community Edition or Enterprise with Console  
When done with the [Prerequisites](#prerequisites), deploy Pachyderm on your local cluster by following these steps:  
  
* Get the Repo Info:  
   ```shell  
   helm repo add pach https://helm.pachyderm.com  
   helm repo update 
   ```  
 * Install Pachyderm:  

=== "Pachyderm Community Edition"
      This command will install Pachyderm's latest available GA version.

       ```shell  
       helm install pachd pach/pachyderm --set deployTarget=LOCAL  
       ```    
=== "Enterprise (Console)"
       - Create a `license.txt` file in which you paste your [Enterprise Key](../../enterprise){target=_blank}.
       - Then, run the following helm command to **install Pachyderm's latest version with Console**: 
       
       ```shell  
       helm install pachd pach/pachyderm --set deployTarget=LOCAL  --set pachd.enterpriseLicenseKey=$(cat license.txt) --set console.enabled=true  
       ``` 

!!! Warning  "Deploying locally with Console requires an **Enterprise Key**"
     * To request a FREE trial enterprise license key, [click here](../../enterprise){target=_blank}. 
     * We create a default mock user (username:`admin`, password: `password`) to [authenticate to Console](../../deploy-manage/deploy/console/#connect-to-console){target=_blank} without having to connect your Identity Provider. 

!!! Tip "To uninstall Pachyderm fully"
      Running `helm uninstall pachd` leaves persistent volume claims behind. To wipe your instance clean, run:
      ```shell
      helm uninstall pachd 
      kubectl delete pvc -l suite=pachyderm 
      ```

!!! Info "See Also"
      More [details on Pachyderm's Helm installation](../../deploy-manage/deploy/helm-install/){target=_blank}.


## Check Your Install

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
At a very minimum, you should see the following pods (console depends on your choice above): 

```shell
NAME                                           READY   STATUS      RESTARTS   AGE
pod/console-5b67678df6-s4d8c                   1/1     Running     0          2m8s
pod/etcd-0                                     1/1     Running     0          2m8s
pod/pachd-c5848b5c7-zwb8p                      1/1     Running     0          2m8s
pod/pg-bouncer-7b855cb797-jqqpx                1/1     Running     0          2m8s
pod/postgres-0                                 1/1     Running     0          2m8s
```

If you see a few restarts on the `pachd` nodes, that means that
Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
Kubernetes restarted those pods. Re-run `kubectl get pods`
 

## Connect 'pachctl' To your Cluster

Assuming your `pachd` is running as shown above, the easiest way to connect `pachctl` to your local cluster is to use the `port-forward` command.

- To connect to your new Pachyderm instance, run:

      ```shell
      pachctl config import-kube local --overwrite
      pachctl config set active-context local
      ```

- Then:

      ```shell
      pachctl port-forward
      ``` 
      **Background this process in a new tab of your terminal.**

=== "You have deployed Enterprise/Console"

        - To connect to your Console (Pachyderm UI), point your browser to **`localhost:4000`** 
        and authenticate using the mock User (username: `admin`, password: `password`).

        - Alternatively, you can connect to your Console (Pachyderm UI) directly by
        pointing your browser to port `4000` on your minikube IP (run `minikube ip` to retrieve minikube's external IP) or docker desktop IP **`http://<dockerDesktopIdaddress-or-minikube>:4000/`** 
        then authenticate using the mock User (username: `admin`, password: `password`).

        - To use `pachctl`, you will need to run `pachctl auth login` then
        authenticate again (to Pachyderm this time) with the mock User (username: `admin`, password: `password`).


* Verify that `pachctl` and your cluster are connected. 
  
    ```shell  
    pachctl version  
    ```  
  
    **System Response:**  
  
    ```  
    COMPONENT           VERSION  
    pachctl             {{ config.pach_latest_version }}  
    pachd               {{ config.pach_latest_version }}  
    ```  
    You are all set!  

  
## Next Steps  
  
Complete the [Beginner Tutorial](../beginner-tutorial) to learn the basics of Pachyderm, such as adding data to a repository and building analysis pipelines.  
  
  
!!! note "See Also:"  
    [General Troubleshooting](../troubleshooting/general-troubleshooting.md)  
  














