# Overview

When you deploy a Kubernetes application, like Pachyderm, typically, it cannot
be accessed from outside the Kubernetes cluster right away. This ensures that
your cluster is secure and resilient to malicious cyber attacks. In the
simplest example, such as running a Pachyderm cluster locally, implicit and
explicit port-forwarding enables you to communicate with `pachd`, the Pachyderm
daemon pod, and `pachd-dash`, the Pachyderm UI. Port-forwarding can be used in
cloud environments as well, but a production environment might require you to
define more sophisticated inbound connection rules.

Kubernetes provides multiple ways to deliver external traffic to a service,
including the following:

* Through a Service with `type: NodePort`. A NodePort service provides basic
access to services. By default, the `pachd` service is deployed as `NodePort`
to simplify interaction with your localhost. `NodePort` is a limited solution
that might not be considered reliable or secure enough in production
deployments. 

* Through a Service with `type: LoadBalancer`. A Kubernetes service with
`type: LoadBalancer` can perform basic load balancing. Typically, if you
change the `pachd` service type to LoadBalancer in a cloud provider, the
cloud provider automatically deploys a load balancer to serve your
application. The downside of this approach is that you will have to change
all your services to the load balancer type and have a separate load
balancers for each service. This can become difficult to manage long term.

* Through the `Ingress` Kubernetes resource. An ingress resource is
completely independent of the services that you deploy. Because an Ingress
resource provides advanced routing capabilities, it is the recommended option
to use with Pachyderm. The only complication of this approach is that you need
to deploy an ingress controller, such as NGINX or traefik, in your Kubernetes
cluster. Such an ingress controller is not deployed by default with the
`pachctl deploy command`.

## Pachyderm Ingress Requirements

Kubernetes supports multiple ingress controller options, and you are free to
pick the one that works best for your environment. However, not all of them
might be fully compatible with Pachyderm. Moreover, exposing your cluster
through an ingress incorrectly might make your Pachyderm cluster and your
data insecure and vulnerable to external attacks. Regardless of your
choice of ingress resource, your environment must meet the
following security requirements to protect your data:

* **Use secure connection**

  Exposing your application to an outside world might pose a security
  risk to your data and organization. Make sure that you have Transport
  Layer Security (TLS) enabled for Ingress connections.

* **Use Pachyderm authentication**

  Pachyderm authentication must be enabled and access provided to a
  verified list of users. Pachyderm authentication is an additional
  security layer to protect your data from malicious attacks.
  If you cannot use Pachyderm authentication providers, we highly recommend to
  use Pachyderm port-forwarding for security reasons. Exposing Pachyderm
  services through an ingress without Pachyderm authentication might result in
  your Pachyderm and Kubernetes clusters being compromised, along with your data.

* **The ingress controller must support gRPC protocol and websockets**

  Some of the ingress controllers that support gRPC include NGNIX and Traefik.

## Ingress Configuration Workflow

This section outlines the general workflow for ingress configuration.
Depending on your use case, you might need to start from the bottom of
this list and determine your firewall and whitelist requirements first.
But commonly, you need to start with deciding which ingress controller
you want to use. In any case, read and understand the requirements
outlined below before you proceed with any configuration.

A general workflow for enabling external traffic inside of a Pachyderm
cluster includes the following steps:

* **Configure Kubernetes networking.**

  You can use one of the following options:

  * (Recommended) Deploy an ingress controller and configure an ingress
  resource.

    Pachyderm supports the [Traefik](https://docs.traefik.io/)
    ingress controller. For more information, see
    [Expose a Pachyderm UI Through an Ingress](../expose-pach-ui-ingress/).

  * Configure the pachd service as a `LoadBalancer` by changing 
  `type: Nodeport` to `type: LoabBalancer` in the `pachd` service
  manifest. As mentioned above, this is the simplest way to expose
  Pachyderm services to the outside world that does not provide
  any sophisticated control over load balancing. This option works
  on most cloud platforms, such as AWS and GKE, as well as in
  minikube, and majorly used for internal use.

* **Configure access to your ingress public IP addresses through firewalls
and whitelisting.**

  If you are deploying Pachyderm on a cloud provider, you need to make sure
  that the ingress IP is available to external users. For example, in AWS, you can
  configure access through security groups in the Virtual Private Cloud (VPC)
  on which the Kubernetes with Pachyderm runs. Other cloud providers have
  similar functionality.

* **Secure the connection end-to-end.**

  If you run Pachyderm in a cloud platform, the cloud provider is responsible
  for securing the underlying infrastructure, such as the Kubernetes control
  plane. Most cloud providers have a security compliance program that address
  these issues. If you are running Kubernetes locally, the security of
  Kubernetes APIs, kubelet, and other components becomes your responsibility.
  See security recommendations in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/). 

  As for Pachyderm, you need to make sure that you deploy Pachyderm with
  TLS enabled. You can deploy `pachd` and `dash` with different certificates
  if required. Self-signed certificates might require additional configuration.
  For instructions on deployment with TLS, see [Deploy Pachyderm with TLS](https://docs.pachyderm.com/latest/deploy-manage/deploy/deploy_w_tls/).

  In addition, you must have administrative access to the Domain Name
  Server (DNS) that you will use to access Pachyderm. If you are deploying
  Pachyderm to an internal site with a self-signed certificate, contact our
  support organization for assistance.

!!! note "See Also"

    - [Expose a Pachyderm UI Through an Ingress](../expose-pach-ui-ingress/)
