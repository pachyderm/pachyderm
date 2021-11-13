# Expose the Pachyderm UI Through an Ingress Resource

Before completing the steps in this section, read [Overview](../configure-external-access/).

A Kubernetes Ingress enables you to allow external traffic inside
of a Pachyderm cluster. Specifically, if you do not want to use
port-forwarding, you might want to expose the Pachyderm UI, `dash`
through an Ingress resource. To do so, you need to deploy an ingress
controller and ingress resource object on the Kubernetes cluster that
runs your Pachyderm cluster. While Kubernetes supports multiple ingress
controllers, not all of them might work seamlessly with Pachyderm.
See [Pachyderm Ingress Requirements](../configure-external-access/#pachyderm-ingress-requirements/) for more information.

In addition, follow these recommendations:

* Follow the recommendations in [the Traefik Kubernetes installation documentation](https://docs.traefik.io/v1.7/user-guide/kubernetes/#deploy-traefik-using-a-deployment-or-daemonset)  that are appropriate for your situation
for installing Traefik.
* You can deploy the `Ingress` resource in any namespace.
* You need to have administrative access to the hostname that you
specify in the `hostname` field in the `Ingress` resource `.yaml`.
* Do not create routes (`paths`) with `pachd` services or S3 services
in the `Ingress` resource `.yaml`.

This section provides an example of how to deploy a Traefik ingress
controller in minikube.

!!! note
    If you cannot use Traefik, you can try to deploy another ingress
    controller, such as NGINX, using the provided example.

To expose the Pachyderm UI through a Traefik resource, complete
the following steps:

1. Deploy a Pachyderm cluster to a cloud provider of your choice as
described in [Deploy Pachyderm](../../deploy/).
For production deployments, deploy with the `--tls` flag.
1. Enable authentication as described in [Configure Access Controls](../../../enterprise/auth/auth/).
1. Deploy the Traefik ingress controlleri as described in the [Traefik Documentation](https://docs.traefik.io/v1.7/user-guide/kubernetes/):

   **Note:** The commands below deploy an example Traefik ingress.
   For the latest version of Traefik controller, see the
   Traefik documentation.

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/pachyderm/pachyderm/1.13.x/examples/traefik-ingress/roles.yaml
   kubectl apply -f https://raw.githubusercontent.com/pachyderm/pachyderm/1.13.x/examples/traefik-ingress/traefik-daemonset.yaml
   ```

1. Deploy the Ingress resource. The following text is an example
metadata file. You need to adjust it to your deployment:

   <script src="https://gist.github.com/svekars/9c665571b665bbd999e050cb829adf2e.js"></script>

   Make sure that you modify the following parameters:

   * `hostname` — an address at which the Pachyderm UI will be
   available. For example, `example.com`.
   * `port` — the ports on which the `grpc-proxy` and the `dash`
   application will run. It can be the same or different ports.

   When done editing, save the file and run:

   ```shell
   kubectl apply -f traefik-ingress-minikube.yaml
   ```

   After you deploy the Traefik ingress, the Pachyderm service
   should be available at the `<hostname>:<port>` that you have
   specified.
