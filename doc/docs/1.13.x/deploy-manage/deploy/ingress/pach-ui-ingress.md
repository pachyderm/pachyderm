# Set up Ingress with Traefik to access Pachyderm UI (`dash`) service in your cluster 
Before completing the following steps, read the [Overview](../index).
This section provides an example of how to route
cluster-external requests (URLs - hostname and path) to cluster-internal services
(here Pachyderm UI (`dash`) service) 
using the **ingress controller** Traefik.
 
However, we recommend you choose the ingress controller
implementation that best fits your cluster.
For Pachyderm UI (`dash`) service in particular,
make sure that it supports WebSockets (Traefik, Nginx, Ambassador...).

Pachyderm UI requires two ports to be open:

- An HTTP port used to serve static assets 
(HTML, CSS, JS, images, etc; port 30080 by default routes to the dash container within the dash pod)
- A WebSocket port, used to connect to the GRPC proxy 
(which in turn connects to pachd; port 30081 by default routes to the grpc-proxy container within the dash pod)  

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
    helm repo add traefik https://helm.traefik.io/traefik
    ```
    ```shell
    helm repo update
    ```

    - Install the Traefik helm chart ([helm v3](https://helm.sh/docs/intro/))
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

1. Create your Ingress route rules (Ingress Ressource):

    The ingress rules expose the private services
    (here Pachyderm UI - `dash` - service) 
    from your cluster to the ingress controller.

    - Create a mypachydermUICRDs.yaml
      ```yaml
      apiVersion: extensions/v1beta1
      kind: Ingress
      metadata:
        name: pachyderm
        annotations:
          kubernetes.io/ingress.class: traefik
          traefik.frontend.rule.type: PathPrefixStrip
      spec:
        rules:
          - host: dash.localhost
            http:
              paths:
                - path: /
                  backend:
                    serviceName: dash
                    servicePort: dash-http
                - path: /ws
                  backend:
                    serviceName: dash
                    servicePort: grpc-proxy-http

      ```

         At a minimum, you will need to specify the following fields:

         - `host` — match the hostname header of the http request (domain).  In the file above,  **dash.localhost** 
         - `path` - mapped to the service port `servicePort` of the `serviceName`. 

   - Create/apply your ressource
      ```shell
      kubectl create -f mypachydermUICRDs.yaml
      ```
      
   - Check your new rules by running `kubectl describe ingress <name-field-value-in-rules-yaml>`:
      ```shell
      kubectl describe ingress pachyderm
      ```
      ```
      Name:             pachyderm
      Namespace:        default
      Address:
      Default backend:  default-http-backend:80 
      Rules:
      Host            Path  Backends
      dash.localhost
                        /     dash:dash-http (10.1.0.7:8080)
                        /ws   dash:grpc-proxy-http (10.1.0.7:8081)
      Annotations:      kubernetes.io/ingress.class: traefik
                        traefik.frontend.rule.type: PathPrefixStrip
      Events:           <none>
      ```
       
   - Check the Traefik Dashboard again (http://127.0.0.1:9000/dashboard/), your new set of rules should now be visible.

!!! Info
       You can deploy your `Ingress` resource (Ingress Route) in any namespace.

!!! Note
       The `servicePort`(s) on which your `serviceName`(s) listens,
       are defined in your Pachyderm deployment manifest (`pachctl deploy local --dry-run > pachd.json`).
       Take a look at the service declaration of `dash` from our manifest below
       and the 2 ports `dash-http` and `grpc-proxy-http` it listens to.
      ```yaml
         "kind": "Service",
         "apiVersion": "v1",
         "metadata": {
            "name": "dash",
            "namespace": "default",
            "creationTimestamp": null,
            "labels": {
               "app": "dash",
               "suite": "pachyderm"
            }
         },
         "spec": {
            "ports": [
               {
               "name": "dash-http",
               "port": 8080,
               "targetPort": 0,
               "nodePort": 30080
               },
               {
               "name": "grpc-proxy-http",
               "port": 8081,
               "targetPort": 0,
               "nodePort": 30081
               }
            ],
      ```      

 
!!! Warning
      - You need to have administrative access to the hostname that you
      specify in the `host` field.
      - Do not create routes (`paths`) with `pachd` or S3 services
      in the `Ingress` resource `.yaml`.
      - Do not create routes (`paths`) with `dash`.
      Pachyderm UI will not load.


## Browse
Connect to your Pachyderm UI: http://dash.localhost/app/. You are all set!

!!! Info
      If you choose to run your WebSocket server on a different host/port, adjust your rules accordingly.
      ```shell
      rules:
      - host: dash.localhost
         http:
            paths:
            - path: /
            backend:
               serviceName: dash
               servicePort: dash-http
      - host: dashws.localhost
         http:
            paths:
            - path: /ws
            backend:
               serviceName: dash
               servicePort: grpc-proxy-http
      ``` 
       You can then access Pachyderm UI by specifying the path, host, and port for the WebSocket proxy
       using GET parameters in the url: http://dash.localhost/app?host=dashws.localhost&path=ws&port=80
       If you’re using the same hostname on your ingress to map both the WebSocket port and the UI port,
       you can omit those parameters.


## References
* [Traefik](https://doc.traefik.io/traefik/v1.7/user-guide/kubernetes/) documentation.
* Kubernetes [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/).
* Kubernetes [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/).



