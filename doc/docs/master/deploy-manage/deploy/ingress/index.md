# Overview
In the simplest case, such as running a Pachyderm cluster locally, implicit and
explicit port-forwarding enables you to communicate with `pachd`, the Pachyderm
daemon pod, and `dash`, the Pachyderm UI. Port-forwarding can be used in
cloud environments as well, but a production environment might require you to
define more sophisticated inbound connection rules.

## Deliver external traffic to Pachyderm
Kubernetes provides multiple ways to deliver external traffic to a service,
including:

* `type: NodePort`. A NodePort service provides basic
  access to services. By default, the `pachd` service is deployed as `NodePort`
  to simplify interaction with your localhost. 

!!! Warning
    `NodePort` is a limited solution
    that might not be considered reliable or 
    secure enough in production
    deployments. 

* `type: LoadBalancer`. A Kubernetes service with
  `type: LoadBalancer` (see `pachd` service manifest) perform basic load balancing. 
  Typically, if you set the `pachd` service type to LoadBalancer
  in a cloud provider, the cloud provider automatically
  deploys a load balancer to serve your
  application. 
  The downside of this approach is that you will have to change
  all your services to the load balancer type and have a separate load
  balancer for each service. This can become difficult to manage long-term.
  This option works on most cloud platforms (AWS, GKE...), 
  as well as in minikube, and majorly used for internal use.

* `Ingress` resource. An ingress resource is
  independent of the services that you deploy.
  Because it provides advanced routing capabilities, 
  it is the recommended option to use with Pachyderm. 

!!! Warning
    Kubernetes supports multiple ingress controller options, and you are free to
    pick the one that works best for your environment. 
    However, not all of them
    might be fully compatible with Pachyderm. 
    In particular, it must support gRPC protocol (for `pachd`) and WebSockets (for `dash`)

!!! Info
    No ingress controller is deployed by default with the `pachctl deploy` command.
    **You will need to deploy an ingress
    controller and ingress resource object of your choice** (Nginx, Traefik, Ambassador...) on the Kubernetes cluster that
    runs your Pachyderm cluster.
    See [Expose a Pachyderm UI Through an Ingress](./pach-ui-ingress)


## Pachyderm Ingress Recommendations

If you run Pachyderm in a cloud platform, the cloud provider is responsible
for securing the underlying infrastructure, such as the Kubernetes control plane.
If you are running Kubernetes locally, the security of
Kubernetes APIs, kubelet, and other components become your responsibility.
See security recommendations in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/). 


For production deployments,
regardless of your choice,
we recommend that you:

* **Use a secure connection**

    Make sure that you have Transport
    Layer Security (TLS) enabled for Ingress connections.
    You can deploy `pachd` and `dash` with different certificates
    if required. Self-signed certificates might require additional configuration.
    For instructions on deployment with TLS, 
    see [Deploy Pachyderm with TLS](https://docs.pachyderm.com/latest/deploy-manage/deploy/deploy_w_tls/).

* **Use Pachyderm authentication/authorization**

    Pachyderm authentication is an additional
    security layer to protect your data from unauthorized access.
    See [Configure Access Controls](https://docs.pachyderm.com/latest/enterprise/auth/enable-auth/).

!!! Warning
    Exposing Pachyderm services through an ingress without Pachyderm authentication might result in
    your Pachyderm and Kubernetes clusters being compromised, along with your data.

* **Configure access to your external IP addresses through firewalls or your Cloud Provider Network Security.**

    In addition, you must have administrative access to the Domain Name
    Server (DNS) that you will use to access Pachyderm. 


  

