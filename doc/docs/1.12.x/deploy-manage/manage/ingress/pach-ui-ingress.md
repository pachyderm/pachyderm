# Set up Ingress with Traefic to access Pachyderm UI (`dash`) service in your cluster 

Before completing the steps in this section, read [Overview](../index).


This section provides an example of how to expose
Pachyderm UI (`dash`) service through Traefik's ingress controller. 
It is designed to give a quick understanding of 
an ingress controller set up for Pachyderm UI.

Feel free to use any other Ingress Controller
that supports websockets (Nginx, Ambassador...).

!!! Warning 

    These installation steps are for **Informational Purposes** ONLY. 
    Please refer to your full Traefic documentation 
    for further installation details and any troubleshooting advice.
   

## Traefic ingress controller set up on Pachyderm's cluster node: Overview
![pach-ui-ingress](../pach-ui-ingress.png)


## Traefic installation and Ingress Route CRD creation
1. Prerequisite:
   * A Pachyderm cluster  - See [Deploy Pachyderm](../../../deploy/).
   * For production deployments, deploy with the [`--tls` flag](https://docs.pachyderm.com/latest/deploy-manage/deploy/deploy_w_tls/). Authentication enableas described in [Configure Access Controls](../../../../enterprise/auth/enable-auth/).
1. Helm install [Traefic](https://github.com/traefik/traefik-helm-chart):

    - Get Repo Info
    ```shell
    helm repo add traefik https://helm.traefik.io/traefik
    helm repo update
    ```

    - Install the Traefic helm chart ([helm v3](https://helm.sh/docs/intro/))
    ```shell
    $ helm install traefik traefik/traefik
    ```

   - Run a quick check:
    ```shell
    $ kubectl get all 
    ```
    then access your Traefic Dashboard at http://127.0.0.1:9000/dashboard/ following the port-forward instructions (You can choose to apply your own Ingress Route CRD instead.):
    ```shell
    $ kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name) 9000:9000
    ```

1. Create your Ingress Route CRDs for Pachyderm UI service (`dash`) in Kubernetes:

 The ingress rules expose the private services from your cluster to the ingress controller.

    - Create a mypachydermUICRDs.yaml
       ```shell
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
         * `host` — here,  **dash.localhost** - match the hostname header of the http request.
         * `path` - mapped to the service port `servicePort` of the `serviceName`. 

      !!! Warning 
      * You need to have administrative access to the hostname that you
      specify in the `host` field.
      * Do not create routes (`paths`) with `pachd` or S3 services
      in the `Ingress` resource `.yaml`.
      * Do not create routes (`paths`) with `dash`.
      Pachyderm UI will not load.


      Look for the `serviceName` and `servicePort` from your Ingress Route in your Pachyderm deployment manifest (`pachctl deploy local --dry-run > pachd.json`) to find the ports on which your service listens.

      ```shell
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
   
    - Create your Ingress Route CRDs:
       ```shell
       $ kubectl create -f mypachydermUICRDs.yaml
       ```
      !!! Note
            You can deploy your `Ingress` resource (Ingress Route) in any namespace.

		 Check the Traefic Dashboard again (http://127.0.0.1:9000/dashboard/), you should now see your new rules.
       
## Browse
Connect to your Pachyderm UI: http://dash.localhost/app/

!!! Note: 
      If you choose to run your websocket server on a different host/port, adjust your rules accordingly.
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
      The access Pachyderm UI by specifying the path, host, and port for the websocket proxy
      using GET parameters in the url: http://dash.localhost/app?host=dashws.localhost&path=ws&port=80
      If you’re using the same hostname on your ingress to map both the websocket port and the UI port, you can omit that last parameter.

## References
* [Traefic](https://doc.traefik.io/traefik/v1.7/user-guide/kubernetes/) documentation.
* [Kubernetes](https://kubernetes.io/docs/concepts/services-networking/ingress/) Ingress.

