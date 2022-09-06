# Deploy Pachyderm With a Proxy: One Port For All External Traffic

We are now shipping Pachyderm with an **optional embedded proxy** allowing Pachyderm to expose one single port externally (whether you access `pachd` over gRPC using `pachctl`, or `console` over HTTP, for example).

See Pachyderm new high-level architecture diagram:
![High level architecture](../images/arch-diagram-high-level-with-proxy.svg)

This page is an add-on to existing installation instructions in the case where you chose to deploy Pachyderm with an embedded proxy. The steps below replace all or parts of the existing installation documentation. We will let you know when to use them and which section they overwrite.

!!! Note "TL;DR" 
    - When the proxy option is activated, Pachyderm is reachable through **one TCP port for all incoming grpc (grpcs), console (HTTP/HTTPS), s3 gateway, OIDC, and dex traffic**, then routes each call to the appropriate backend microservice without any additional configuration.
    - Enable the proxy as follow:

      ```yaml
      proxy:
        enabled: true
        service:
          type: LoadBalancer
      ```

!!! Warning
    The deployment of Pachyderm with a proxy is optional at the moment and will become permanent in the next minor release of Pachyderm.

The diagram below gives a quick overview of the layout of services and pods when using a proxy. In particular, it details how Pachyderm listens to all inbound traffic on one port, then routes each call to the appropriate backend:![Infrastruture Recommendation](../images/infra-recommendations-with-proxy.png)

!!! Note 
    See our [reference values.yaml](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/pachyderm/values.yaml#L827){target=_blank} for all available configurable fields of the proxy.

Before any deployment in production, we recommend reading the following section to [set up your production infrastructure](#deploy-pachyderm-in-production-with-a-proxy). 

Alternatively, you can skip those infrastructure prerequisites and make a [quick cloud installation](#quick-cloud-deployment-with-a-proxy) or jump to our [local deployment](#deploy-pachyderm-locally-with-a-proxy) section for a first encounter with Pachyderm.

## Pachyderm General Infrastructure Recommendations

For production deployments, we recommend that you:

* **Provision a TCP load balancer** for all HTTP/HTTPS, gRPC/gRPCs, aws s3, /dex incoming traffic.
The TCP load balancer (load balanced at L4 of the OSI model) will have port `80/443` forwarding to the `pachyderm-proxy` service entry point. Please take a look at the diagram above.
        
    When a proxy is enabled with `type:LoadBalancer` (see the snippet of values.yaml enabling the proxy), Pachyderm creates a `pachyderm-proxy` service allowing your cloud platform (AWS, GKE...) to **provision a TCP Load Balancer automatically**.
        
    !!! Note 
        - You can optionally attach any additional Load Balancer configuration information to the metadata of your service by adding the appropriate `annotations` in the `proxy.service` of your values.yaml.
        - You can pre-create a static IP (For example, in GCP: `gcloud compute addresses create ADDRESS_NAME --global --IP-version IPV4)`, then pass this external IP to the `loadBalancerIP` in the `proxy.service` of your values.yaml.

        ```yaml
        proxy:
          enabled: true
          service:
            type: LoadBalancer
            annotations: {<add-optional-annotations-here}
            loadBalancerIP: <insert-your-proxy-external-IP-address-here>
        ```

* **Use a secure connection**

    Make sure that you have [Transport Layer Security (TLS)](../deploy-w-tls){target=_blank} enabled for your incoming traffic. 
   
* **Use Pachyderm authentication/authorization**

    Pachyderm authentication is an additional
    security layer to protect your data from unauthorized access.
    See the [authentication and authorization section](../../../enterprise/auth/) to activate access control and set up an Identity Provider (IdP).

* **Configure access to your external IP addresses through firewalls or your Cloud Provider Network Security.**

* (Optional) **Create a DNS entry for your public IP**

## Deploy Pachyderm in Production With a Proxy

Once you have your networking infrastructure setup, 
check the [deployment page that matches your cloud provider](../../){target=_blank} and 
follow the installation steps that apply to the cloud provider of your choice from section 1-6.
Make sure that you have enabled the proxy by adding the following lines to your values.yaml:

```yaml
proxy:
  enabled: true
  service:
    type: LoadBalancer
    annotations: {see examples below}
```

Once your cluster is provisioned, and Pachyderm installed, 
replace the instructions in section 7 (Have 'pachctl' And Your Cluster Communicate) by this [new set of instructions](#to-connect-your-pachctl-client-to-your-cluster).

!!! Attention "If you plan to deploy Console in Production, read the following and adjust your values.yaml accordingly."
    Deploying Pachyderm with a proxy simplifies the setup of Console (No more dedicated DNS and ingress needed in front of Console). In a production environment, you will need to:

    - [Activate Authentication](../../../enterprise/auth/){target=_blank}.Although, if you are an Helm user, setting up your License Key in your values.yaml will activate Authentication by default. This instruction applies to users activating auth by using pachctl.
    - Update the values in the highlighted fields below.
    - Additionally, you will need to configure your Identity Provider (`oidc.upstreamIDPs`). See examples for the `oidc.upstreamIDPs` value in the [helm chart values specification](https://github.com/pachyderm/pachyderm/blob/42462ba37f23452a5ea764543221bf8946cebf4f/etc/helm/pachyderm/values.yaml#L461){target=_blank} and read [our IDP Configuration page](../../../enterprise/auth/authentication/idp-dex) for a better understanding of each field. 

    ```yaml hl_lines="10-11 19-20"

    deployTarget: "<pick-your-cloud-provider>"

    # enable the proxy
    proxy:
      enabled: true
      service:
        type: LoadBalancer
        annotations: {...}

    ingress:
      host: <insert-external-ip-address-or-dns-name>

    pachd:
      storage:
        amazon:
          bucket: "<bucket-name>"
          ...
          region: "<us-east-2>"
      # pachyderm enterprise key
      enterpriseLicenseKey: "<your-enterprise-token>"

    oidc:
      # populate the pachd.upstreamIDPs with an array of Dex Connector configurations.
      upstreamIDPs: []
    ```

### To connect your `pachctl` client to your cluster
The grpc address provided when pointing your `pachctl` CLI at your cluster changes now that a proxy allows a single entry point.
Run the following commands:

1. Retrieve the external IP address of your TCP load balancer (or use your domain name):
  ```shell
  kubectl get services | grep pachyderm-proxy | awk '{print $4}'
  ```
1. Update the context of your cluster using the external IP address/domain name captured above:

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

  Your cluster context name should show up. Your `pachctl` client now points to your cluster.

### If you have deployed **Console**
Point your browser to `http://<external-IP-address-or-domain-name>`. No port number is needed. You will be prompted to log in to your Console.

### If you have installed **JupyterHub and the Mount Extension**
The connection string to your Pachyderm cluster (check the login form accessible by clicking on the mount extension icon in the far left tab bar of your JupyterLab) now depends on whether you have deployed JupyterHub on:

- The same cluster: `grpc://pachd.<namespace>.svc.cluster.local:30650`
- An external cluster: `grpc://<external-IP-address-or-domain-name>:80`

## Quick Cloud Deployment With a Proxy

Follow your regular [QUICK Cloud Deploy documentation](../quickstart/), but for those few steps:

- In section 2 (Create Your Values.yaml), replace your values yaml with the YAML files provided below. Make sure to replace the dummy values with their relevant information. Then proceed with the helm installation as detailed in section 3. 
- To [connect your `pachctl` client to your cluster](#to-connect-your-pachctl-client-to-your-cluster),replace section 4 with the instructions detailed in the link.
- To [connect to Console](#if-you-have-deployed-console), replace section 5 with the instructions provided in the link.
- If you deployed JupyterHub (section 7), use the instructions in the link to [login to the Mount Extension](#if-you-have-installed-jupyterhub-and-the-mount-extension). 


### AWS
=== "Deploy Pachyderm without Console"

    ```yaml hl_lines="3-6"
    deployTarget: "AMAZON"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    pachd:
      storage:
        amazon:
          bucket: "bucket_name"      
          # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
          id: "AKIAIOSFODNN7EXAMPLE"                
          # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
          region: "us-east-2"          
    ```
=== "Deploy Pachyderm with Console and Enterprise"

    ```yaml hl_lines="3-9 21-26"
    deployTarget: "AMAZON"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    ingress:
      host: <insert-external-ip-address-or-dns-name>

    pachd:
      storage:
        amazon:
          bucket: "<bucket-name>"                
          # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
          id: "AKIAIOSFODNN7EXAMPLE"                
          # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
          region: "<us-east-2>"
      # pachyderm enterprise key 
      enterpriseLicenseKey: "<your-enterprise-token>"
      localhostIssuer: "true"
    ```

### Google

=== "Deploy Pachyderm without Console"

    ```yaml hl_lines="3-6"
    deployTarget: "GOOGLE"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    pachd:
      storage:
        google:
          bucket: "<bucket-name>"
          cred: |
            INSERT JSON CONTENT HERE
      externalService:
        enabled: true
    ```
=== "Deploy Pachyderm with Console and Enterprise"

    ```yaml hl_lines="3-9 18-19"
    deployTarget: "GOOGLE"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    ingress:
      host: <insert-external-ip-address-or-dns-name>

    pachd:
      storage:
        google:
          bucket: "<bucket-name>"
          cred: |
            INSERT JSON CONTENT HERE
      # pachyderm enterprise key
      enterpriseLicenseKey: "<your-enterprise-token>"
      localhostIssuer: "true"
    ```

### Azure

=== "Deploy Pachyderm without Console"

    ```yaml hl_lines="3-6"
    deployTarget: "MICROSOFT"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    pachd:
      storage:
        microsoft:
          # storage container name
          container: "blah"
          # storage account name
          id: "AKIAIOSFODNN7EXAMPLE"
          # storage account key
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    ```
=== "Deploy Pachyderm with Console and Enterprise"

    ```yaml hl_lines="3-9 22-23"
    deployTarget: "MICROSOFT"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

    ingress:
      host: <insert-external-ip-address-or-dns-name>


    pachd:
      storage:
        microsoft:
          # storage container name
          container: "<your-container-name>"
          # storage account name
          id: "AKIAIOSFODNN7EXAMPLE"
          # storage account key
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      # pachyderm enterprise key
      enterpriseLicenseKey: "<your-enterprise-token>"
      localhostIssuer: "true"
    ```
## Deploy Pachyderm Locally With a Proxy

This section is an alternative to the default [local deployment instructions](../../../getting-started/local-installation){target=_blank}. It uses a variant of the original one line command to enable a proxy. 

Follow the [Prerequisites](#prerequisites){target=_blank} before [deploying Pachyderm](#deploy-pachyderm-community-edition-or-enterprise) (with or without Console) on your local cluster, then [Connect 'pachctl' To Your Cluster](#connect-pachctl-to-your-cluster).

JupyterLab users, [**you can also install Pachyderm JupyterLab Mount Extension**](../../how-tos/jupyterlab-extension/index.md#pachyderm-jupyterlab-mount-extension){target=_blank} on your local Pachyderm cluster to experience Pachyderm from your familiar notebooks. 

Note that you can run both Console and JupyterLab on your local installation.
### Prerequisites

- If you are not using Linux, follow all the default [Prerequisites](../../../getting-started/local-installation/#prerequisites){target=_blank} installation instructions. 

- If you are a Linux user, make sure to set up your local Kubernetes Cluster with [Kind](../../../getting-started/local-installation/#using-kind){target=_blank} while following the default [Prerequisites](../../../getting-started/local-installation/#prerequisites){target=_blank} installation instructions. **Use the command below**.

Then start your Kubernetes environment.

=== "Minikube (OS X / Windows)"

    ```shell
    minikube start
    ```

    Later, we will use `minikube tunnel` to make the proxy available on `localhost`.

    Check [Minikube's documentation](https://minikube.sigs.k8s.io/docs/){target=_blank} for details.

=== "Kind (Linux)"

    ```shell
      cat <<EOF | kind create cluster --name=kind --config=-
      kind: Cluster
      apiVersion: kind.x-k8s.io/v1alpha4
      nodes:
          - role: control-plane
            kubeadmConfigPatches:
                - |
                    kind: InitConfiguration
                    nodeRegistration:
                        kubeletExtraArgs:
                            node-labels: "ingress-ready=true"
            extraPortMappings:
                - containerPort: 30080
                  hostPort: 80
                  protocol: TCP
                - containerPort: 30443
                  hostPort: 443
                  protocol: TCP
        EOF
    ```

      The extraPortMappings will make NodePorts in the cluster available on localhost; NodePort 30080 becomes localhost:80.
      This will make Pachyderm available at `localhost:80` as long as this kind cluster is running.

      Check [Kind's documentation](https://kind.sigs.k8s.io/){target=_blank} for details.

### Deploy Pachyderm Community Edition Or Enterprise

* Get the Repo Info:  

  ```shell  
  helm repo add pach https://helm.pachyderm.com  
  helm repo update 
  ```  


* Install Pachyderm by running the following command:  
 

!!! Attention "Attention Kind users"
    
     Set your Service type to `NodePort` rather than `LoadBalancer` in the commands below.

      ```shell 
      -- set proxy.service.type=NodePort
      ```
          
=== "Community Edition With Console"

      ```shell
      helm install pachd pach/pachyderm --set deployTarget=LOCAL --set proxy.enabled=true --set proxy.service.type=LoadBalancer 
      ```
=== "Enterprise With Console"

      This command will unlock your enterprise features and install Console Enterprise. Note that Console Enterprise requires authentication. By default, we create a default mock user (username:admin, password: password) to authenticate to Console without having to connect your Identity Provider.

      - Create a license.txt file in which you paste your [Enterprise Key](../../../enterprise/){target=_blank} .
      - Then, run the following helm command to install Pachyderm's latest Enterprise Edition:

        ```shell 
        helm install pachd pach/pachyderm --set deployTarget=LOCAL --set proxy.enabled=true --set proxy.service.type=LoadBalancer --set pachd.enterpriseLicenseKey=$(cat license.txt) --set ingress.host=localhost
        ```
    


* Check Your Install

Check the status of the Pachyderm pods by periodically
running `kubectl get pods`. When Pachyderm is ready for use,
all Pachyderm pods must be in the **Running** status.


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

### Connect 'pachctl' To Your Cluster

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

- To use `pachctl`, run `pachctl auth login` then
authenticate again (to Pachyderm this time) with the mock User (username: `admin`, password: `password`).

- Notebook users, if you have installed [JupyterHub and the Mount Extension](../../how-tos/jupyterlab-extension/index.md){target=_blank} on the same cluster, the connection url to your Pachyderm cluster in the login form (click on the mount extension icon in the far left tab ) is now: `grpc://pachd.<namespace>.svc.cluster.local:30650`

## Changes to the S3 Gateway

The `pachyderm-proxy` service also routes Pachyderm's [**S3 gateway**](../../manage/s3gateway/){target=_blank} (allowing you to
**access Pachyderm's repo through the S3 protocol**) on port 80 (note the endpoint in the diagram below).

![Global S3 Gateway with Proxy](../../images/main-s3-gateway-with-proxy.png)

## Changes to the Enterprise Server Setup

Your enterprise server is [deployed in the same way as any regular cluster](#deploy-pachyderm-in-production-with-a-proxy) with a few differences (no object-store and two PostgreSQL databases required: `dex` and `pachyderm`). The same applies when deploying an enterprise server with a proxy. 

Note that the enterprise server will be deployed behind its proxy, as will each cluster registered to this enterprise server.

!!! Attention
    Enabling an embedded enterprise server with your pachd as part of the same helm installation will not work with the proxy. 
    You can use a standalone enterprise server instead.

Follow your regular [enterprise server deployment and configuration instructions](../../../enterprise/auth/enterprise-server/setup){target=_blank}, except for those few steps:

- [Section 1: Deploy an enterprise server](../../enterprise/auth/enterprise-server/setup.md#1-deploy-an-enterprise-server)):
   
    In the values.yaml provided as examples:

    - Remove the `pachd.externalService` section and replace it with `proxy`:

        ```yaml
        proxy:
          enabled: true
          service:
            type: LoadBalancer
        ```

    - Update all mentions of `http://<PACHD-IP>:30657/` and `http://<PACHD-IP>:30658/` with `http://<Enterprise-server-external-IP-or-DNS>:80/` or `https://<Enterprise-server-external-IP-or-DNS>:443/`

    - Your `redirect_uri` must be set to `http(s)://<insert-external-ip-or-dns-name>/dex/callback` in your IdP connector as mentioned in the [IdP section of the documentation](../../../enterprise/auth/authentication/idp-dex/#pachyderm-integration-with-identity-providers){target=_blank}


- [Section 3: Register your cluster with the enterprise server](../../enterprise/auth/enterprise-server/setup.md#3-register-your-cluster-with-the-enterprise-server){target=_blank}:

    If you chose to [register a cluster to an enterprise server using pachctl](../../enterprise/auth/enterprise-server/setup.md#register-clusters-with-pachctl){target=_blank}, change all the port numbers to 80(http)/443(https) in the `pachctl enterprise register` command:

    ```shell
    pachctl enterprise register --id <my-pachd-config-name> --enterprise-server-address <pach-enterprise-IP>:80 --pachd-address <pachd-IP>:80
    ```

- [Section 4: Enable auth on each cluster](../../enterprise/auth/enterprise-server/setup.md#4-enable-auth-on-each-cluster){target=_blank}, use these instructions to:

    - Set up the issuer in the idp config between the enterprise server and your cluster:
    
    ```shell
    echo "issuer: http://<enterprise-external-IP-or-dns>" | pachctl idp set-config --config -
    ```

    - For each registered cluster, enable auth:

        ```shell
        pachctl auth activate --client-id <my-pachd-config-name> --redirect http://<pachd-external-IP-or-DNS>/authorization-code/callback 
        ```

    - Then resume the last part of instructions:

        - Make sure than your enterprise context is set up properly:

        ```shell
        pachctl config get active-enterprise-context
        ```

        If not:

        ```shell
        pachctl config set active-enterprise-context <my-enterprise-context-name>
        ```





