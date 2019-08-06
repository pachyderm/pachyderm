# Connect to a Pachyderm cluster

After you deploy a Pachyderm cluster you can continue to use
the command-line interface, connect to the UI, or configure
third-party applications to access your cluster programmatically.
Often all you need to do is just continue to use the command-line
interface to create repositories, pipelines, and upload your data.
At other times, you might have multiple Pachyderm deployed and need
to switch between them to perform management operations.

You do not need to configure anything specific to start using the
Pachyderm CLI right after deployment. However, the Pachyderm
dashboard and the S3 gateway functionality require
explicit port-forwarding.

This section describes the various options available for you to connect
to your Pachyderm cluster.

## Connect by using a Pachyderm context

You can specify an IP address that you use to connect to the
Pachyderm UI and the S3 gateway by storing that address in the
Pachyderm configuration file as the `pachd_address` parameter.
If you have already deployed a Pachyderm cluster, you can
set a Pachyderm IP address by updating you cluster configuration
file.

This configuration is supported for deployments that do not have
a firewall set up between the Pachyderm and the client machine.
Defining a dedicated `pachd` IP address is a more reliable and
secure way compared to port-forwarding. Therefore, Pachyderm
recommends that you use contexts in all production environments.

For more information about Pachyderm contexts, see
[Manage Cluster Access](../managing_pachyderm/connect-to-cluster.html)

To connect by using a Pachyderm context, complete the following
step:

1. Get the current context:

   ```bash
   $ pachctl config get active-context
   ```

   If no IP address is set up for this cluster, you get the following
   output:

   ```bash
   $ pachctl config get context <name>
   {

   }
   ```

1. Set up `pachd_address`:

   ```bash
   $ pachctl config update context <name> --pachd-address <host:port>
   ```

   **Example:**

   ```bash
   $ pachctl config update context local --pachd-address 10.10.10.15:30650
   ```

   **Note:** The `pachd` port is `30650` or `650`.

1. Verify that the configuration has been updated:

   ```bash
   $ pachctl config get context local
   {
      "pachd_address": "10.10.10.15:30650"
   }
   ```

1. Use the new IP address to connect to your Pachyderm cluster

## Connect by using port-forwarding

The Pachyderm dashboard is the Pachyderm user interface that enables you
to manage Pachyderm resources, create pipelines, view visual
representation of your pipeline directed acyclic graph (DAG), and
perform other operations.

Because Pachyderm dashboard runs in a Kubernetes pod, you cannot access
it without forwarding connections from your local network port to a
network pod in the dashboard pod by explicitly running Kubernetes
port-forwarding on your machine.

To enable port-forwarding, open a new terminal window and run
`pachctl port-forward`.

## Connect by using the VM IP address

You can connect to the Pachyderm dashboard by using the VM IP address
with the port `:30080`. 


