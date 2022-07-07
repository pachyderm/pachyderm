# Set up Ingress with Traefik to access Pachyderm UI (`console`) service in your cluster 

Before completing the following steps, read the [Infrastructure Recommendation page](../).

This section provides an example of how to route
cluster-external HTTP/HTTPS requests to cluster-internal services
(here Pachyderm UI `console` service and `authentication` services
using the **ingress controller** Traefik.
 

## Traefik ingress controller on Pachyderm UI's cluster in one diagram
Here is a quick high-level view of the various components at play.
![pach-ui-ingress](../../../images/console_ingress_traefik.png)

!!! Warning 

    The following installation steps are for **Informational Purposes** ONLY. 
    Please refer to your full Traefik documentation 
    for further installation details and any troubleshooting advice.

## Traefik installation and Ingress Ressource Definition
1. Helm install [Traefik](https://github.com/traefik/traefik-helm-chart){target=_blank}:

    - Get Repo Info
    ```shell
    helm repo add traefik https://helm.traefik.io/traefik
    ```
    ```shell
    helm repo update
    ```

    - Install the Traefik helm chart ([helm v3](https://helm.sh/docs/intro/){target=_blank})
    ```shell
    helm install traefik traefik/traefik
    ```

   - Run a quick check:
    ```shell
    kubectl get all 
    ```
    You should see your Traefik pod, service, deployments.apps, and replicaset.app.

    You can now access your Traefik Dashboard at http://127.0.0.1:9000/dashboard/ following the port-forward instructions (You can choose to apply your own Ingress Ressource instead.):
    ```shell
    kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name) 9000:9000
    ```

1. Configure the Ingress in the helm chart.
   You will need to configure any specific annotations your ingress controller requires. 

    - my_pachyderm_values.yaml
       ```yaml
       ingress:
         enabled: false
         annotations:
            kubernetes.io/ingress.class: "traefik"
            traefik.ingress.kubernetes.io/router.tls: "true"
         host: "<your_domain_name>"
       ```

       For a list of all available annotations, read the [Traefik & Kubernetes documentation](https://doc.traefik.io/traefik/routing/providers/kubernetes-ingress/){target=_blank}.

       At a minimum, you will need to specify the `host` field: match the hostname header of the http request (domain).  


       Check the [list of all available helm values](https://github.com/pachyderm/pachyderm/blob/42462ba37f23452a5ea764543221bf8946cebf4f/etc/helm/pachyderm/values.yaml#L143){target=_blank} at your disposal in our reference documentation.

1. Install Pachyderm and Console using the Helm Chart

      Once you have your [networking infrastructure](./) set up, apply a helm values file such as the one specified in the example file below to wire up routing through an Ingress, and set up TLS.


      ```yaml
      ingress:
         enabled: true
         host: <DNS-ENTRY-A>
         annotations:
            ## annotations specific to integrate with your ingress-controller
            traefik.ingress.kubernetes.io/router.tls: "true"
            kubernetes.io/ingress.class: "traefik"
         tls:
            enabled: true
            secretName: "pach-tls"
      pachd:
         tls:
            enabled: true
            secretName: "pach-tls"
         externalService:
            enabled: true
            loadBalancerIP: <DNS-ENTRY-B>
      console:
         enabled: true
      ```
      ```shell
      helm install pachd -f my_pachyderm_values.yaml pach/pachyderm
      ```
      The deployment of Pachyderm automatically creates the required set of rules.

1. Check your new rules by running `kubectl describe ingress console`:
         ```shell
         kubectl describe ingress console
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
       
1. Check the Traefik Dashboard again (http://127.0.0.1:9000/dashboard/){target=_blank}, your new set of rules should now be visible.


## Browse
Connect to your Console (Pachyderm UI): `https://<external-IP-address-or-domain-name>:443/` (if TLS is enabled) or `http://<external-IP-address-or-domain-name>:80/`. You are all set!

## References
* [Traefik](https://doc.traefik.io/traefik/v1.7/user-guide/kubernetes/){target=_blank} documentation.




