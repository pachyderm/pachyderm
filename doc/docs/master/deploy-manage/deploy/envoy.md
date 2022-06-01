# One Port For All External Traffic

!!! Warning
    Envoy is an [experimental feature](../../../contributing/supported-releases/#experimental).

We are now shipping Pachyderm with an **optional embedded proxy ([Envoy](https://www.envoyproxy.io/))** so users only need to expose one port to the Internet (wether they access `pachd` over gRPC using `pachctl`, or `console` over HTTP, for example to).

!!! Note "TL;DR" 
    - When envoy is activated, Pachyderm is reachable through **one TCP port for all incoming grpc (grpcs), console (http/https), s3 gateway, OIDC, and dex traffic**, then routes each call to the appropriate backend microservice without any additional configuration.
    - To deploy Pachyderm with Envoy, enable the proxy as follow:

    ```yaml
    proxy:
      enabled: true
      service:
        type: LoadBalancer
    ```
    
The high-level architecture diagram below gives a quick overview of the layout of services and pods when using Envoy. In particular, it details how Pachyderm serves all external traffic on one port, then routes each call to the appropriate backend:
![Infrastruture Recommendation](../../images/infra-recommendations-envoy.png)

!!! Note
    - See our [reference values.yaml](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml#L649){target=_blank} for all available configurable fields of the proxy.
    - Refer to our generic ["Helm Install"](../helm-install/) page for more information on how to install and get started with `Helm`.

Read the following recommendations to set up your infrastructure in [production]() or jump to our [local deployment](#deploy-pachyderm-locally-with-envoy) section.

## Pachyderm General Infrastructure Recommendations

For production deployments, we recommend that you:

* **Provision a TCP load balancer** for all HTTP/HTTPS, gRPC/gRPCs, aws s3, /dex incoming traffic.
Provision a TCP load balancer with port `80/443` forwarding to `pachyderm-proxy` service entry point.
      
    In general, for non-local deployments, you should load balance **all gRPC, HTTP, and S3 incoming traffic** to a TCP LB (load balanced at L4 of the OSI model) deployed in front of `pachyderm-proxy` service.
        
    When Envoy is enabled with `type:LoadBalancer`, Pachyderm creates a `pachyderm-proxy` service allowing your cloud platform (AWS, GKE...) to **provision a TCP Load Balancer automatically**.
        
    To provision an external load balancer in your current cloud automatically (if supported), add the appropriate `annotations` in the `proxy` service of your values.yaml (see below) to attach any Load Balancer configuration information to the metadata of your service:

    ```yaml
    proxy:
      enabled: true
      service:
        type: LoadBalancer
        annotations: {see examples below}
    ```


    === "Example on AWS EKS"
        In the following example, we deploy an NLB on AWS EKS:

        ``` yaml
        proxy:
          enabled: true
          service:
            type: LoadBalancer
            annotations: {see examples below}
              service.beta.kubernetes.io/aws-load-balancer-type: "external"
              service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: "ip"
              service.beta.kubernetes.io/aws-load-balancer-scheme: "internal"
              service.beta.kubernetes.io/aws-load-balancer-subnets: "subnet-aaaaa,subnet-bbbbb,subnet-ccccc"

        ```

    === "Example on GCP GKE"
        In the following example, we pre-created a static IP by running `gcloud compute addresses create ADDRESS_NAME --global --ip-version IPV4`, then passed this external IP to the values.yaml as follow:

        ``` yaml
        proxy:
          enabled: true
          service:
            type: LoadBalancer
            loadBalancerIP: ${ADDRESS_NAME}
        ```
    === "Example on Azure AKS"
        This example is identical to the example on Google GKE.

        ``` yaml
          proxy:
            enabled: true
            service:
              type: LoadBalancer
              loadBalancerIP: ${ADDRESS_NAME}
        ```

* **Use a secure connection**

    Make sure that you have Transport
    Layer Security (TLS) enabled for Ingress connections.
    You can deploy `pachd` and `console` with different certificates
    if required. Self-signed certificates might require additional configuration.
    For instructions on deployment with TLS, 
    see [Deploy Pachyderm with TLS](../deploy-w-tls/).

    !!! Note
        Optionally, you can use a certificate manager such as [cert-manager](https://cert-manager.io/docs/){target=_blank} to refresh certificates and inject them as kubernetes secrets into your cluster for the TCP load balancer to use.
   
* **Use Pachyderm authentication/authorization**

    Pachyderm authentication is an additional
    security layer to protect your data from unauthorized access.
    See the [authentication and authorization section](../../../enterprise/auth/) to activate access control and set up an IdP.

* **Configure access to your external IP addresses through firewalls or your Cloud Provider Network Security.**

* (Optional) **Create a DNS entry for your public IP**

## Deploy Pachyderm in Production With Envoy

Once you have your networking infrastructure setup, 
check the [deployment page that matches your cloud provider](../../).

Follow the installation steps applying to your cloud provider. 
The use of Envoy does not change any of the installation steps; However:

- It impacts the grpc address provided when pointing your `pachctl` CLI at your cluster (chapter 7 - Have 'pachctl' And Your Cluster Communicate).

  Run the following commands instead to connect your `pachctl` client to your cluster.

  1. Retrieve the external IP address of your TCP load balancer (or use your domain name):
    ```shell
    kubectl get services | grep pachyderm-proxy | awk '{print $4}'
    ```

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:80"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
      ```
      ```shell
      pachctl config set active-context "<your-cluster-context-name>"
      ```

  1. Check that your are using the right context: 

      ```shell
      pachctl config get active-context
      ```

      Your cluster context name should show up. You are all set.

- If you have deployed Console, point your browser to `<external-IP-address-or-domain-name`. No port number needed. You will be prompted to login to your Console.

- Similarly, if you have installed JupyterHub and the Mount Extension, the connection url to your Pachyderm cluster in the login form accessible by clicking on the mount extension icon in the far left tab bar is now "grpc://<external-IP-address-or-domain-name>:80".
## Deploy Pachyderm Locally With Envoy

This section is an alternative to the default [local deployment instructions](../../../getting-started/local-installation){target=_blank}. It uses a variant of the default local values.yaml enabling the Envoy proxy. 
We will make many references to the original local installation instructions and only highlight the differences.

  
!!! Warning "Reminder"
      - A local installation is **not designed to be a production  
      environment**. It is meant to help you learn and experiment quickly with Pachyderm.   
      - A local installation is designed for a **single-node cluster**.  
      This cluster uses local storage on disk and does not create  
      Persistent Volumes (PVs). If you want to deploy a production multi-node  
      cluster, follow the instructions for your cloud provider or on-prem  
      installation as described in [Deploy Pachyderm](../../deploy-manage/deploy/).  
      New Kubernetes nodes cannot be added to this single-node cluster.   
 

### 0- Prerequisites  
You have installed:  
  
- A [Kubernetes cluster](../../../getting-started/local-installation/#setup-a-local-kubernetes-cluster){target=_blank} running on your local environment (pick the virtual machine of your choice):   
      - [Docker Desktop](../../../getting-started/local-installation/#using-kubernetes-on-docker-desktop){target=_blank} 
      - [Minikube](../../../getting-started/local-installation/#using-minikube){target=_blank}
      - [Kind](../../../getting-started/local-installation/#using-kind){target=_blank}
      - Oracle® VirtualBox™ 
- [Helm](https://helm.sh/docs/intro/install/){target=_blank} to deploy Pachyderm on your Kubernetes cluster.  
- [Pachyderm Command Line Interface (`pachctl`)](../../../getting-started/local-installation/#install-pachctl){target=_blank} to interact with your Pachyderm cluster.
- [Kubernetes Command Line Interface `kubectl`](https://kubernetes.io/docs/tasks/tools/){target=_blank} to interact with your underlying Kubernetes cluster.
 
  
### 1- Deploy Pachyderm Community Edition Or Enterprise
  
When done with the [Prerequisites](#prerequisites), deploy Pachyderm on your local cluster by following these steps. 
Note that the Enterprise version of Pachyderm comes with Console, Pachyderm's UI.

JupyterLab users, [**you can also install Pachyderm JupyterLab Mount Extension**](../../how-tos/jupyterlab-extension/#pachyderm-jupyterlab-mount-extension){target=_blank} on your local Pachyderm cluster to experience Pachyderm from your familiar notebooks. 

Note that you can run both Console and JupyterLab on your local installation.

* Get the Repo Info:  

    ```shell  
    helm repo add pach https://helm.pachyderm.com  
    helm repo update 
    ```  

* Create your values.yaml

=== "Latest CE"

      ```yaml 
      deployTarget: LOCAL

      proxy:
        enabled: true
        service:
          type: LoadBalancer
      ```    
=== "Enterprise (Console)"

    Make sure to update your enterprise key in `pachd.enterpriseLicenseKey`.

      ```yaml 
      deployTarget: LOCAL

      proxy:
        enabled: true
        service:
          type: LoadBalancer
        
      pachd:
        image:
          tag: 2.2.0
        enterpriseLicenseKey: "key"
        oauthRedirectURI: http://localhost/authorization-code/callback
        
      console:
        image:
          tag: 2.2.0-1
        enabled: true
        config:
          reactAppRuntimeIssuerURI: http://localhost
          oauthRedirectURI: http://localhost/oauth/callback/?inline=true
        
      oidc:
        userAccessibleOauthIssuerHost: localhost
      ```
* Install Pachyderm by running the following command:  

  ```shell  
  helm install pachd pach/pachyderm -f values.yaml 
  ```    

!!! Tip "To uninstall Pachyderm fully on Minikube"
      Running `helm uninstall pachd` leaves persistent volume claims behind. To wipe your instance clean, run:
      ```shell
      helm uninstall pachd 
      kubectl delete pvc -l suite=pachyderm 
      ```

### 2- Check Your Install

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
NAME                                  READY   STATUS    RESTARTS   AGE
pod/console-55bc9f679-w4xrk           1/1     Running   0          71m
pod/etcd-0                            1/1     Running   0          70m
pod/pachd-84487d6675-cf68x            1/1     Running   0          71m
pod/pachyderm-proxy-89d5c4f65-pst9l   1/1     Running   0          71m
pod/pg-bouncer-5dd558c8dc-zjlpj       1/1     Running   0          71m
pod/postgres-0                        1/1     Running   0          70m
```

If you see a few restarts on the `pachd` nodes, that means that
Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
Kubernetes restarted those pods. Re-run `kubectl get pods`
 

### 3- Connect 'pachctl' To Your Cluster

Assuming your `pachd` is running as shown above,
you can now connect `pachctl` to your local cluster.

!!! Attention "Minikube users" 
  Open a new tab in your terminal and run `minikube tunnel` (the command creates a network route on your host to `pachyderm-proxy` service deployed with type LoadBalancer, and set its ingress to its ClusterIP, here `127.0.0.1`). You will be prompted to enter your password.

- To connect `pachctl` to your new Pachyderm instance, run:

    ```shell
    echo '{"pachd_address":"grpc://127.0.0.1:80"}' | pachctl config set context local --overwrite && pachctl config set active-context local
    ```

   Verify that `pachctl` and your cluster are connected by running `pachctl version`:
  
    **System Response:**  

    ```  
    COMPONENT           VERSION  
    pachctl             {{ config.pach_latest_version }}  
    pachd               {{ config.pach_latest_version }}  
    ```  
    You are all set!  

- To connect to your Console (Pachyderm UI), point your browser to **`localhost`** (no port number needed)
and authenticate using the mock User (username: `admin`, password: `password`).

- If you have installed JupyterHub and the Mount Extension, pass "grpc://127.0.0.1:80"

- To use `pachctl`, run `pachctl auth login` then
authenticate again (to Pachyderm this time) with the mock User (username: `admin`, password: `password`).

- Notebook users, if you have installed [JupyterHub and the Mount Extension](../../how-tos/jupyterlab-extension/#pachyderm-jupyterlab-mount-extension){target=_blank}, the connection url to your Pachyderm cluster in the login form (click on the mount extension icon in the far left tab ) is now "grpc://<external-IP-address-or-domain-name>:80".

## Next Steps  
  
Complete the [Beginner Tutorial](../beginner-tutorial) to learn the basics of Pachyderm, such as adding data to a repository and building analysis pipelines.  
  
 



























































By default, the local deployment of Pachyderm deploys the `pachyderm-proxy` service as `type:LoadBalancer`.

The plan is:

Make a proxy that serves all traffic on one port available, but optional (proxy.enabled).  Only cleartext HTTP on port 80 for now.  (Currently working in a PR, hopefully part of 2.2.   )

Allow the proxy to serve individual ports without special routing, to be the same as the legacy NodePort/LoadBalancer port.  (proxy.service.type, proxy.service.legacyPorts={ grpc: 30650, ... } (Currently working in a PR, hopefully part of 2.2.)

Allow the proxy to be configured with a TLS certificate and serve secure traffic on port 443, and redirect port 80 to https.  (2.2.something)

Make the proxy the default (2.3)

Make pachd.proxy.enabled default to true.

Make the proxy non-optional (2.4)

Remove the pachd.service and pachd.externalService configuration keys.

Convert any external pachd service in the cluster to point at the proxy instead of directly at pachd.

Consider supporting mTLS; where envoy ↔︎ {pachd, console} traffic is encrypted.  (More user feedback needed on whether or not they need this.) (2.4-ish.)

Depending on the outcome of this, TLS configuration can be removed from pachd and console.  Envoy will terminate TLS, and pachd need not concern itself with updating keys.



https://pachyderm.atlassian.net/browse/ENT-10#icft=ENT-10 and https://pachyderm.atlassian.net/browse/ENT-11#icft=ENT-11 make it possible to just point an Ingress controller, ANY ingress controller, at this Envoy port and have everything work perfectly with no setup.

This epic tracks doing the necessary work to make this work smoothly.  We can remove a lot of hacks that are in place once we have this, specifically relating to local deployment.  Local deployment and real deployment should be identical from a networking perspective now.  I think the only knobs to turn are: what's the external name of this TCP port (needed for oauth redirects), and do you want TLS.