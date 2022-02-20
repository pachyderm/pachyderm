# Connect to a Pachyderm cluster

After you deploy a Pachyderm cluster, you can continue to use
the command-line interface, connect to the Pachyderm dashboard, or
configure third-party applications to access your cluster programmatically.
Often all you need to do is just continue to use the command-line
interface to create repositories, pipelines, and upload your data.
At other times, you might have multiple Pachyderm clusters deployed
and need to switch between them to perform management operations.

You do not need to configure anything specific to start using the
Pachyderm CLI right after deployment. However, the Pachyderm
dashboard and the S3 gateway require
explicit port-forwarding or direct access through an externally
exposed IP address and port.

This section describes the various options available for you to connect
to your Pachyderm cluster.

## Connect to a Local Cluster

If you are just exploring Pachyderm, use port-forwarding to
access both `pachd` and the Pachyderm dashboard.

By default, Pachyderm enables port-forwarding from `pachctl` to `pachd`.
If you do not want to use port-forwarding for `pachctl` operations,
configure a `pachd_address` as described in
[Connect by using a Pachyderm context](#connect-by-using-a-pachyderm-context).

To connect to the Pachyderm dashboard, you can either use port-forwarding,
or the IP address of the virtual machine on which your Kubernetes cluster
is running.

The following example shows how to access a Pachyderm cluster
that runs in `minikube`.

To connect to a Pachyderm dashboard, complete the following steps:

* To use port-forwarding, run:

  ```shell
  pachctl port-forward
  ```

* To use the IP address of the VM:`

  1. Get the minikube IP address:

     ```shell
     minikube ip
     ```

  1. Point your browser to the following address:

     ```shell
     <minikube_ip>:30080
     ```

## Connect to a Cluster Deployed on a Cloud Platform

As in a local cluster deployment, a Pachyderm cluster
deployed on a cloud platform has impllcit port-forwarding enabled.
This means that, if you are connected to the cluster so
that `kubectl` works, `pachctl` can communicate with `pachd`
without any additional configuration.
Other services still need explicit port forwarding.
For example, to access the Pachyderm dashboard,
you need to run pachctl port-forward.
Since a Pachyderm cluster
deployed on a cloud platform is more likely to become
a production deployment, configuring a `pachd_address`
as described in
[Connect by using a Pachyderm context](#connect-by-using-a-pachyderm-context)
is the preferred way.

## Connect by using a Pachyderm context

You can specify an IP address that you use to connect to the
Pachyderm UI and the S3 gateway by storing that address in the
Pachyderm configuration file as the `pachd_address` parameter.
If you have already deployed a Pachyderm cluster, you can
set a Pachyderm IP address by updating your cluster configuration
file.

This configuration is supported for deployments that do not have
a firewall set up between the Pachyderm cluster deployed in the cloud
and the client machine.
Defining a dedicated `pachd` IP address and host is a more reliable
way that might also result in better performance compared to
port-forwarding. Therefore, Pachyderm
recommends that you use contexts in all production environments.

This configuration requires that you deploy an ingress controller
and a reliable security method to protect the traffic between the
Pachyderm pods and your client machine. Remember that exposing your
traffic through a public ingress might
create a security issue. Therefore, if you expose your `pachd` endpoint,
you need to make sure that you take steps to protect the endpoint and
traffic against common container security threats. Port-forwarding
might be an alternative which might result in sufficient performance
if you place the data that is consumed by your pipeline in object
store buckets located in the same region.

For more information about Pachyderm contexts, see
[Manage Cluster Access](../manage/cluster-access.md).

To connect by using a Pachyderm context, complete the following
steps:

1. Get the current context:

   ```shell
   pachctl config get active-context
   ```

   This command returns the name of the currently active context.
   Use this as the argument to the command below.

   If no IP address is set up for this cluster, you get the following
   output:

   ```shell
   pachctl config get context <name>
   {

   }
   ```

1. Set up `pachd_address`:

   ```shell
   pachctl config update context <name> --pachd-address <host:port>
   ```

   **Example:**

   ```shell
   pachctl config update context local --pachd-address 192.168.1.15:30650
   ```

   **Note:** By default, the `pachd` port is `30650`.

1. Verify that the configuration has been updated:

   ```shell
   pachctl config get context local
   ```

   **System Response:**

   ```shell
   {
      "pachd_address": "192.168.1.15:30650"
   }
   ```

## Connect by Using Port-Forwarding

The Pachyderm port-forwarding is the simplest way to enable cluster access
and test basic Pachyderm functionality. Pachyderm automatically starts
port-forwarding from `pachctl` to `pachd`. Therefore, the traffic
from the local machine goes to the `pachd` endpoint through the
Kubernetes API. However, to open a persistent tunnel to other ports, including
the Pachyderm dashboard, git and authentication hooks, the built-in HTTP
file API, and other, you need to run port-forwarding explicitly.

Also, if you are connecting with port-forward, you are using the `0.0.0.0`.
Therefore, if you are using a proxy, it needs to handle that appropriately.

Although port-forwarding is convenient for testing, for production
environments, this connection might be too slow and unreliable.
The speed of the port-forwarded traffic is limited to `1 MB/s`.
Therefore, if you experience high latency with `put file` and
`get file` operations, or if you anticipate high throughput
in your Pachyderm environment, you need to enable ingress access
to your Kubernetes cluster. Each cloud provider has its own
requirements and procedures for enabling ingress controller and
load balancing traffic at the application layer for better performance.

Remember that exposing your traffic through a public ingress might
create a security issue. Therefore, if you expose your `pachd` endpoint,
you need to make that you take steps to protect the endpoint and
traffic against common container security threats. Port-forwarding
might be an alternative which might provide satisfactory performance
if you place the data that is consumed by your pipeline in object
store buckets located in the same region.

<!--Add a link to the section that describes the above-->

To enable port-forwarding, complete the following steps:

1. Open a new terminal window.
1. Run:

   ```shell
   pachctl port-forward
   ```

   This command does not stop unless you manually interrupt it.
   You can run other `pachctl` commands from another window.
   If any of your `pachctl` commands hang, verify if the
   `kubectl` port-forwarding has had issues that prevent
   `pachctl port-forward` from running properly.
