# Deploy Console

!!! Important  
    To deploy Pachyderm console,
    an [***Enterprise License***](../../../enterprise/) is required. 


Note that this section is an add-on to the deployment of Pachyderm, locally or in the cloud. 
It details the additional steps required to install and access your Console.

- If you plan to deploy in the cloud, find Pachyderm's deployment instructions matching your target (AWS, Google, Azure...) in the [Deploy section](../) of the documentation, update your values.yaml accordingly, then complement with the following instructions.
- To deploy locally, follow the instructions below.

## Deploy Locally

!!! Info "Reminder"
      A local installation helps you learn
      some of the Pachyderm basics and experiment with the product. It is not designed to be a production environment.

We provide an easy "one line" deployment command to install pachyderm with Console on a local environment. All you need is your enterprise token.

Run the following helm installation:

```shell
$ helm install pachd pach/pachyderm --set deployTarget=LOCAL --set pachd.activateEnterprise=true --set pachd.enterpriseLicenseKey=$(cat license.txt) --set console.enabled=true
```

!!! Note
     By default, a mock Identity Provider will be set up with username & password set to, `admin` &`password` respectively. 

To connect to your Console (Pachyderm UI):

* Enable port forwarding by running `pachctl port-forward`, and then point your browser to
`localhost:4000`.
* Alternatively, you can connect to your Console (Pachyderm UI) directly by pointing your
    browser to port `4000` on your minikube IP (run `minikube ip` to retrieve minikube's external IP) or docker desktop IP `http://<dockerDesktopIdaddress-or-minikube>:4000/`.  
* Authenticate using `admin` & `password`. 

You are all set!

!!! Note
    Note that you can set up an IDP locally (Okta, Auth0...) and authenticate as one of its users. You can choose to do so when deploying Pachyderm (in the values.yaml) or later on (using pachctl). Either way, follow the steps described in the `production environment` section below.
## Deploy In The Cloud

The deployment of Console in your favorite Cloud requires, at a minimum, the set up an [Ingress](../ingress/#ingress). 


Once your Ingress is configured in your values.yaml, you can choose one of the following:

- For a **quick installation** of Console (Not recommended in Production but an easy way to get started), set up a **mockIDP** during the deployment of Pachyderm by providing the following fields in your values.yaml then [connect to your Console](#connect-to-console):

    ```yaml
    pachd.activateEnterprise = true
    pachd.enterpriseLicenseKey = <LICENSE>
    oidc.mockIDP = true
    console.enabled = true
    ```

!!! Note
    - By default, the mock Identity Provider will be set up with username & password set to `admin` &`password`, respectively. 
    - You can set up an Identity Provider later on by using `pachctl`, see [our IDP Configuration document](../../../enterprise/auth/authentication/idp-dex).

- For a **production environment**:

    - Set up your IDP during deployment.
        To configure your Identity Provider as a part of `helm install`, see examples for the `oidc.upstreamIDPs` value in the [helm chart values specification](https://github.com/pachyderm/pachyderm/blob/42462ba37f23452a5ea764543221bf8946cebf4f/etc/helm/pachyderm/values.yaml#L461) and read [our IDP Configuration document](../../../enterprise/auth/authentication/idp-dex) for a better understanding of each field. 
    - Or manually remove the mockIDP (i.e., update your values.yaml with `oidc.mockIDP = false`) then [set up an Identity Provider by using `pachctl`](../../../enterprise/auth/authentication/idp-dex).



!!! Warning
    - **When enterprise is enabled through Helm, auth is automatically activated** (i.e., you do not need to run `pachctl auth activate`) and a `pachyderm-bootstrap-config` k8s secret is created containing an entry for your [rootToken](../../../enterprise/auth/#activate-user-access-management). Use `kubectl get secret pachyderm-bootstrap-config -o yaml` to retrieve it and save it where you see fit.

    However, **this secret is only used when configuring through helm**:

     - If you run `pachctl auth activate`, the secret is not updated. Instead, the rootToken is printed in your STDOUT for you to save.
     - Same behavior if you [activate enterprise manually](../../../enterprise/deployment/) (`pachctl license activate`) then [activate authentication](../../../enterprise/auth/) (`pachctl auth activate`).
 
## Connect to Console
To connect to your Console (Pachyderm UI):

- Point your browser to:

    - `http://<external-IP-address-or-domain-name>:80` or,
    - `https://<external-IP-address-or-domain-name>:443` if TLS is enabled

- Authenticate:

    - As the mock User using `admin` & `password` if you used the mockIDP.
    - As a User of your IdP otherwise.

You are all set! 
You should land on the Projects Page of Console.

![Console Landing Page](../../../getting_started/images/console_landing_page.png)

!!! Info "See Also"
      More [details on Pachyderm's Helm installation](../../deploy-manage/deploy/helm_install/).

