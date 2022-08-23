# Deploy Console

!!! Important  
    To deploy Pachyderm's Console,
    an [***Enterprise License***](../../../enterprise/) is required. 


Note that this section is an add-on to the deployment of Pachyderm, locally or in the cloud. 
It details the additional steps required to install and access your Console.

- If you plan to deploy in the cloud, this section will complement your values.yaml (find Pachyderm's deployment instructions matching your target ([AWS](../aws-deploy-pachyderm/), [Google](../google-cloud-platform/), [Azure](../azure/)...) in the [Deploy section](../) of the documentation).
- To deploy locally, follow the instructions below.

## Deploy Locally

!!! Info "Reminder"
      A local installation helps you learn
      some of the Pachyderm basics and experiment with the product. It is not designed to be a production environment.

We provide an easy "one line" deployment command to install Pachyderm with Console on a local environment. All you need is your enterprise token and [a Kubernetes cluster running on your local environment](../../../getting-started/local-installation/#prerequisites).

Follow the deployment instructions in our [Local Installation](../../getting-started/local-installation#deploy-pachyderm-community-edition-or-enterprise-with-console) page.
You are all set!

!!! Note
    When installing, we create a default mock user (username:`admin`, password: `password`) to authenticate to Console without the hassle of connecting your Identity Provider.

## Deploy In The Cloud

The deployment of Console in your favorite Cloud usually requires, at a minimum, the set up an Ingress (see below), the activation of Authentication, and the setup of a DNS.

- You can opt for a **quick installation** that will alleviate those infrastructure constraints (Not recommended in Production but an easy way to get started) and speed up your installation by following the steps in our [Quick Cloud Deployment](../quickstart/) page, then [connect to your Console](#connect-to-console): 

!!! Note "Reminder"
    - Use the mock user (username:`admin`, password: `password`) to authenticate to Console.

- For a **production environment**:

    - Set up your [Ingress](../ingress/#ingress) and DNS.
    - Set up your IDP during deployment.
        To configure your Identity Provider as a part of `helm install`, see examples for the `oidc.upstreamIDPs` value in the [helm chart values specification](https://github.com/pachyderm/pachyderm/blob/42462ba37f23452a5ea764543221bf8946cebf4f/etc/helm/pachyderm/values.yaml#L461){target=_blank} and read [our IDP Configuration page](../../../enterprise/auth/authentication/idp-dex) for a better understanding of each field. 
    - Or manually update your values.yaml with `oidc.mockIDP = false` then [set up an Identity Provider by using `pachctl`](../../../enterprise/auth/authentication/idp-dex).

!!! Warning
    - **When enterprise is enabled through Helm, auth is automatically activated** (i.e., you do not need to run `pachctl auth activate`) and a `pachyderm-bootstrap-config` k8s secret is created containing an entry for your [rootToken](../../../enterprise/auth/#activate-user-access-management). Use `{{"kubectl get secret pachyderm-bootstrap-config -o go-template='{{.data.rootToken | base64decode }}'"}}` to retrieve it and save it where you see fit.
    
        However, **this secret is only used when configuring through helm**:

        - If you run `pachctl auth activate`, the secret is not updated. Instead, the rootToken is printed in your STDOUT for you to save.
        - Same behavior if you [activate enterprise manually](../../../enterprise/deployment/) (`pachctl license activate`) then [activate authentication](../../../enterprise/auth/) (`pachctl auth activate`).


    - **Set the helm value `pachd.activateAuth` to false to prevent the automatic bootstrap of auth on the cluster**.

## Connect to Console

=== "No Ingress set up (Local or Quick Install)"

    - Run `pachctl port-forward` (Background this process in a new tab of your terminal).
    
    - Connect to your Console (Pachyderm UI):

         - Point your browser to `http://localhost:4000` 
         - Authenticate as the mock User using `admin` & `password` 

=== "Ingress + DNS set up"

    - Point your browser to:

         - `http://<external-IP-address-or-domain-name>:80` or,
         - `https://<external-IP-address-or-domain-name>:443` if TLS is enabled

    - Authenticate:

         - As the mock User using `admin` & `password` if you used the mockIDP.
         - As a User of your IdP otherwise.


You are all set! 
You should land on the Projects page of Console.

![Console Landing Page](../../../getting-started/images/console_landing_page.png)

