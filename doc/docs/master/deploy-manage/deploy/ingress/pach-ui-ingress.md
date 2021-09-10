# Set up Ingress with Traefik to access Pachyderm UI (`console`) service in your cluster 
Before completing the following steps, read the [Overview](../index).
This section provides an example of how to route
cluster-external requests (URLs - hostname and path) to cluster-internal services
(here Pachyderm UI (`console`) service 
using the **ingress controller** Traefik.
 
However, we recommend you choose the ingress controller
implementation that best fits your cluster.

Pachyderm UI requires a single (HTTP) port to be open.

## Traefik ingress controller on Pachyderm UI's cluster in one diagram
Here is a quick high-level view of the various components at play.
![pach-ui-ingress](../pach-ui-ingress.png)

!!! Warning 

    The following installation steps are for **Informational Purposes** ONLY. 
    Please refer to your full Traefik documentation 
    for further installation details and any troubleshooting advice.

## Traefik installation and Ingress Ressource Definition
1. Helm install [Traefik](https://github.com/traefik/traefik-helm-chart):

    - Get Repo Info
    ```shell
    $ helm repo add traefik https://helm.traefik.io/traefik
    ```
    ```shell
    $ helm repo update
    ```

    - Install the Traefik helm chart ([helm v3](https://helm.sh/docs/intro/))
    ```shell
    $ helm install traefik traefik/traefik
    ```

   - Run a quick check:
    ```shell
    $ kubectl get all 
    ```
    You should see your Traefik pod, service, deployments.apps, and replicaset.app.

    You can now access your Traefik Dashboard at http://127.0.0.1:9000/dashboard/ following the port-forward instructions (You can choose to apply your own Ingress Ressource instead.):
    ```shell
    $ kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name) 9000:9000
    ```

1. Configure the Ingress in the helm chart.
   You will need to configure any specific annotations your ingress controller requires. 

    - my_pachyderm_values.yaml
      ```yaml
      ingress:
         enabled: false
         annotations:
            kubernetes.io/ingress.class: traefik
            traefik.frontend.rule.type: PathPrefixStrip
         host: "console.localhost"
      ```

         At a minimum, you will need to specify the following fields:

         - `host` — match the hostname header of the http request (domain).  In the example above,  **console.localhost** 

!!! Note
      Check the [list of all available helm values](../../../../reference/helm_values/) at your disposal in our reference documentation.

   - Install Pachyderm using the Helm Chart
      ```shell
      $ helm install pachd -f my_pachyderm_values.yaml pach/pachyderm
      ```
      
   - Check your new rules by running `kubectl describe ingress console`:
      ```shell
      $ kubectl describe ingress console
      ```
      ```
      Name:             console
      Namespace:        default
      Address:
      Default backend:  default-http-backend:80 
      Rules:
      Host            Path  Backends
      console.localhost
                        /     console:console-http (10.1.0.7:4000)
      Annotations:      kubernetes.io/ingress.class: traefik
                        /dex     pachd:identity-port (10.1.0.8:1658)
      Annotations:      kubernetes.io/ingress.class: traefik
                        /     pachd:oidc-port (10.1.0.8:1657)
      Annotations:      kubernetes.io/ingress.class: traefik
      Events:           <none>
      ```
       
   - Check the Traefik Dashboard again (http://127.0.0.1:9000/dashboard/), your new set of rules should now be visible.


!!! Warning
      - You need to have administrative access to the hostname that you
      specify in the `host` field.
      - Do not create routes (`paths`) with `pachd` or S3 services
      in the `Ingress` resource `.yaml`.
      - Do not create routes (`paths`) with `console`.
      Pachyderm UI will not load.


## Browse
Connect to your Pachyderm UI: http://console.localhost/app/. You are all set!

## References
* [Traefik](https://doc.traefik.io/traefik/v1.7/user-guide/kubernetes/) documentation.
* Kubernetes [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/).
* Kubernetes [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/).



