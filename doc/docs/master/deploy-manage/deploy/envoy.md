# Deploy Pachyderm With Envoy: One Port For All External Traffic

We are now shipping Pachyderm with an **optional embedded proxy ([Envoy](https://www.envoyproxy.io/))** allowing Pachyderm to expose one single port to the Internet (whether you access `pachd` over gRPC using `pachctl`, or `console` over HTTP, for example).

This page is an add-on to existing installation instructions in the case where you chose to enable Envoy when deploying Pachyderm. The steps below replace all or parts of the existing installation documentation. We will let you know when to use them and which section they overwrite.

!!! Note "TL;DR" 
    - When Envoy is activated, Pachyderm is reachable through **one TCP port for all incoming grpc (grpcs), console (HTTP/HTTPS), s3 gateway, OIDC, and dex traffic**, then routes each call to the appropriate backend microservice without any additional configuration.
    - To deploy Pachyderm with Envoy, enable the proxy as follow:

    ```yaml
    proxy:
      enabled: true
      service:
        type: LoadBalancer
    ```

!!! Warning
    Envoy is an optional deployment option that will become permanent in the next minor release of Pachyderm.

The high-level architecture diagram below gives a quick overview of the layout of services and pods when using Envoy. In particular, it details how Pachyderm listens to all inbound traffic on one port, then routes each call to the appropriate backend:![Infrastruture Recommendation](../../images/infra-recommendations-envoy.png)

!!! Note 
    See our [reference values.yaml](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml#L649){target=_blank} for all available configurable fields of the proxy.

Before any deployment in production, we recommend reading the following recommendations section and [set up your production infrastructure](#deploy-pachyderm-in-production-with-envoy). 

Alternatively, you can skip those infrastructure prerequisites and make a [quick cloud installation](#quick-cloud-deployment-with-envoy) or jump to our [local deployment](#deploy-pachyderm-locally-with-envoy) section for the first encounter with Pachyderm.

## Pachyderm General Infrastructure Recommendations

For production deployments, we recommend that you:

* **Provision a TCP load balancer** for all HTTP/HTTPS, gRPC/gRPCs, aws s3, /dex incoming traffic.
The TCP load balancer (load balanced at L4 of the OSI model) will have port `80/443` forwarding to the `pachyderm-proxy` service entry point. Please take a look at the diagram above.
        
    When Envoy is enabled with `type:LoadBalancer` (see the snippet of values.yaml enabling the proxy), Pachyderm creates a `pachyderm-proxy` service allowing your cloud platform (AWS, GKE...) to **provision a TCP Load Balancer automatically**.
        
    To provision this external load balancer automatically (if supported) and attach any Load Balancer configuration information to the metadata of your servic, add the appropriate `annotations` in the `proxy.service` of your values.yaml (see below):

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

    Make sure that you have Transport Layer Security (TLS) is enabled for your incoming traffic. If required, you can deploy `pachd` and `console` with different certificates. Self-signed certificates might require additional configuration. For instructions on deployment with TLS, see [Deploy Pachyderm with TLS](../deploy-w-tls/).

    !!! Note 
        Optionally, you can use a certificate manager such as [cert-manager](https://cert-manager.io/docs/){target=_blank} to refresh certificates and inject them as Kubernetes secrets into your cluster for the TCP load balancer to use.
   
* **Use Pachyderm authentication/authorization**

    Pachyderm authentication is an additional
    security layer to protect your data from unauthorized access.
    See the [authentication and authorization section](../../../enterprise/auth/) to activate access control and set up an Identity Provider (IdP).

* **Configure access to your external IP addresses through firewalls or your Cloud Provider Network Security.**

* (Optional) **Create a DNS entry for your public IP**

## Deploy Pachyderm in Production With Envoy

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
    Deploying Pachyderm with Envoy simplifies the setup of Console (No more dedicated DNS and ingress needed in front of Console). In a production environment, you will need to:

    - Activate Authentication.
    - Update the values in the highlighted fields below.
    - Additionally, you will need to configure your Identity Provider (`oidc.upstreamIDPs`). See examples for the `oidc.upstreamIDPs` value in the [helm chart values specification](https://github.com/pachyderm/pachyderm/blob/42462ba37f23452a5ea764543221bf8946cebf4f/etc/helm/pachyderm/values.yaml#L461){target=_blank} and read [our IDP Configuration page](../../../enterprise/auth/authentication/IDP-dex) for a better understanding of each field. 

    ```yaml hl_lines="18 19-29"

    deployTarget: "<pick-your-cloud-provider>"

    # enable the proxy
    proxy:
      enabled: true
      service:
        type: LoadBalancer
        annotations: {...}

    pachd:
      storage:
        amazon:
          bucket: "<bucket-name>"
          ...
          region: "<us-east-2>"
      # pachyderm enterprise key
      enterpriseLicenseKey: "<your-enterprise-token>"
      oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/authorization-code/callback

    console:
      enabled: true
      config:
        reactAppRuntimeIssuerURI: http://<insert-external-ip-address-or-dns-name>
        oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/oauth/callback/?inline=true

    oidc:
      userAccessibleOauthIssuerHost: <insert-external-ip-address-or-dns-name>
      # populate the pachd.upstreamIDPs with an array of Dex Connector configurations.
      upstreamIDPs: []
    ```

### To connect your `pachctl` client to your cluster
The grpc address provided when pointing your `pachctl` CLI at your cluster changes now that Envoy allows a single entry point.
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
The connection string to your Pachyderm cluster (check the login form accessible by clicking on the mount extension icon in the far left tab bar of your JupyterLab) is now `grpc://<external-IP-address-or-domain-name>:80`.

## Quick Cloud Deployment With Envoy

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
=== "Deploy Pachyderm with Console"

    ```yaml hl_lines="3-6 18-30"
    deployTarget: "AMAZON"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

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
      oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/authorization-code/callback

    console:
      enabled: true
      config:
        reactAppRuntimeIssuerURI: http://<insert-external-ip-address-or-dns-name>
        oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/oauth/callback/?inline=true
    
    oidc:
      userAccessibleOauthIssuerHost: <insert-external-ip-address-or-dns-name>
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
=== "Deploy Pachyderm with Console"

    ```yaml hl_lines="3-6 15-26"
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
      # pachyderm enterprise key
      enterpriseLicenseKey: "<your-enterprise-token>"
      localhostIssuer: "true"
      oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/authorization-code/callback

    console:
      enabled: true
      config:
        reactAppRuntimeIssuerURI: http://<insert-external-ip-address-or-dns-name>
        oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/oauth/callback/?inline=true

    oidc:
      userAccessibleOauthIssuerHost: <insert-external-ip-address-or-dns-name>
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
=== "Deploy Pachyderm with Console"

    ```yaml hl_lines="3-6 18-29"
    deployTarget: "MICROSOFT"

    proxy:
      enabled: true
      service:
        type: LoadBalancer

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
      oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/authorization-code/callback

    console:
      enabled: true
      config:
        reactAppRuntimeIssuerURI: http://<insert-external-ip-address-or-dns-name>
        oauthRedirectURI: http://<insert-external-ip-address-or-dns-name>/oauth/callback/?inline=true

    oidc:
      userAccessibleOauthIssuerHost: <insert-external-ip-address-or-dns-name>
    ```
## Deploy Pachyderm Locally With Envoy

This section is an alternative to the default [local deployment instructions](../../../getting-started/local-installation){target=_blank}. It uses a variant of the original local values.yaml to enable Envoy. 

When done with the [Prerequisites](../../../getting-started/local-installation/#prerequisites){target=_blank}, [deploy Pachyderm](#deploy-pachyderm-community-edition-or-enterprise-with-console) (with or without Console) on your local cluster by using the following values.yaml rather than the one provided in the original installation steps, then [Connect 'pachctl' To Your Cluster](#connect-pachctl-to-your-cluster).

JupyterLab users, [**you can also install Pachyderm JupyterLab Mount Extension**](../../how-tos/jupyterlab-extension/#pachyderm-jupyterlab-mount-extension){target=_blank} on your local Pachyderm cluster to experience Pachyderm from your familiar notebooks. 

Note that you can run both Console and JupyterLab on your local installation.

### Deploy Pachyderm Community Edition Or Enterprise with Console 

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
        enterpriseLicenseKey: "key"
        localhostIssuer: "true"
        oauthRedirectURI: http://localhost/authorization-code/callback
        
      console:
        enabled: true
        config:
          reactAppRuntimeIssuerURI: http://localhost
          oauthRedirectURI: http://localhost/oauth/callback/?inline=true
        
      oidc:
        mockIDP: true
        userAccessibleOauthIssuerHost: localhost
      ```
* Install Pachyderm by running the following command:  

  ```shell  
  helm install pachd pach/pachyderm -f values.yaml 
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

- If you have installed JupyterHub and the Mount Extension, pass "grpc://127.0.0.1:80"

- To use `pachctl`, run `pachctl auth login` then
authenticate again (to Pachyderm this time) with the mock User (username: `admin`, password: `password`).

- Notebook users, if you have installed [JupyterHub and the Mount Extension](../../how-tos/jupyterlab-extension/#pachyderm-jupyterlab-mount-extension){target=_blank}, the connection url to your Pachyderm cluster in the login form (click on the mount extension icon in the far left tab ) is now 'grpc://<external-IP-address-or-domain-name>:80'.

  