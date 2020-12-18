# Migrate to a Minor or Major Version

!!! info
    If you need to upgrade Pachyderm from one patch
    to another, such as from x.xx.0 to x.xx.1, see
    [Upgrade Pachyderm](upgrades.md).

As new versions of Pachyderm are released, you might need to update your
cluster to get access to bug fixes and new features.

Migrations involve moving between major releases, such as 1.x.x to
2.x.x or minor releases, such as 1.9.x to 1.10.0.

!!! tip
    Pachyderm follows the [Semantic Versioning](https://semver.org/)
    specification to manage the release process.

Pachyderm stores all of its states in the following places:

* In `etcd` which in turn stores its state in one or more persistent volumes,
which were created when the Pachyderm cluster was deployed. `etcd` stores
metadata about your pipelines, repositories, and other Pachyderm primitives.

* In an object store bucket, such as AWS S3, MinIO, or Azure Blob Storage.
Actual data is stored here.

In a migration, the data structures stored in those locations need to be
read, transformed, and rewritten. Therefore, this process involves the
following steps:

1. Back up your cluster by exporting the existing Pachyderm cluster's repos,
pipelines, and input commits to a backup file and optionally to an S3 bucket.
1. Bring up a new Pachyderm cluster adjacent to the old pachyderm cluster either
in a separate namespace or in a separate Kubernetes cluster.
1. Restore the old cluster's repos, commits, and pipelines into the new
   cluster.

!!! warning
    Whether you are upgrading or migrating your cluster, you must back it up
    to guarantee that you can restore it after migration.

## Step 1 - Back up Your Cluster

Before migrating your cluster, create a backup that you can use to restore your
cluster from. For large amounts of data that are stored in an S3 object store,
we recommend that you use the cloud provider capabilities to copy your data
into a new bucket while backing up information about Pachyderm object to a
local file. For smaller deployments, you can copy everything into a local
file and then restore from that file.

To back up your cluster, complete the following steps:

1. Back up your cluster by running the `pachctl export` command with the
`--no-object` flag as described in [Back up Your Cluster](../backup_restore/).

1. In your cloud provider, create a new S3 bucket with the same Permissions
policy that you assigned to the original cluster bucket. For example,
if your cluster is on EKS, create the same Permissions policy as described
in [Deploy Pachyderm with an IAM Role](../../deploy/amazon_web_services/aws-deploy-pachyderm/#deploy-pachyderm-with-an-iam-role).

1. Clone your S3 bucket that you used for the olf cluster to this new bucket.
   Follow the instructions for your cloud provider:

   * If you use Google cloud, see the [gsutil instructions](https://cloud.google.com/storage/docs/gsutil/commands/cp).
   * If you use Microsoft Azure, see the [azcopy instructions](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).
   * If you use Amazon EKS, see [AWS CLI instructions](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).

   **Example:**

   ```shell
   aws s3 sync s3://mybucket s3://mybucket2
   ```

1. Proceed to [Step 2](#step-2-restore-all-paused-pipelines).

## Step 2 - Restore All Paused Pipelines

If you want to minimize downtime and run your pipeline while you are migrating
your cluster, you can restart all paused pipelines and data loading operations
after the backup and clone operations are complete.

To restore all paused pipelines, complete the following steps:

1. Run the `pachctl start pipeline` command on each paused pipeline or
   use the multi-line shell script to restart all pipelines at once:

   - one-by-one
      ```shell
      pachctl start pipeline <pipeline-name>
      ```

   - all-at-once
      ```shell
      pachctl list pipeline --raw \
      | jq -r '.pipeline.name' \
      | xargs -P3 -n1 -I{} pachctl start pipeline {}
      ```

   You might need to install `jq` and other utilities to run the script.

1. Confirm that each pipeline is started using the `list pipeline` command:

   ```shell
   pachctl list pipeline
   ```

   * If you have switched the ports to stop data loading from outside sources,
   change the ports back:

     1. Back up the current configuration:

        ```shell
        kubectl get svc/pachd -o json >pachd_service_backup_30649.json
        kubectl get svc/etcd -o json >etcd_svc_backup_32379.json
        kubectl get svc/dash -o json >dash_svc_backup_30080.json
        ```

     1. Modify the services to accept traffic on the corresponding ports to
     avoid collisions with the migration cluster:

        ```shell
        # Modify the pachd API endpoint to run on 30650:
        kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kubectl apply -f -
        # Modify the pachd trace port to run on 30651:
        kubectl get svc/pachd -o json | sed 's/30648/30651/g' | kubectl apply -f -
        # Modify the pachd api-over-http port to run on 30652:
        kubectl get svc/pachd -o json | sed 's/30647/30652/g' | kubectl apply -f -
        # Modify the pachd SAML authentication port to run on 30654:
        kubectl get svc/pachd -o json | sed 's/30646/30654/g' | kubectl apply -f -
        # Modify the pachd git API callback port to run on 30655:
        kubectl get svc/pachd -o json | sed 's/30644/30655/g' | kubectl apply -f -
        # Modify the pachd s3 port to run on 30600:
        kubectl get svc/pachd -o json | sed 's/30611/30600/g' | kubectl apply -f -
        # Modify the etcd client port to run on 32378:
        kubectl get svc/etcd -o json | sed 's/32378/32379/g' | kubectl apply -f -
        # Modify the dashboard ports to run on 30081 and 30080:
        kubectl get svc/dash -o json | sed 's/30079/30080/g' | kubectl apply -f -
        kubectl get svc/dash -o json | sed 's/30078/30081/g' | kubectl apply -f -
        ```

1. Modify your environment so that you can access `pachd` on the old port:

   ```shell
   pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650
   ```

1. Verify that you can access `pachd`:

   ```shell
   pachctl version
   ```

   **System Response:**

   ```
   COMPONENT           VERSION
   pachctl             1.10.0
   pachd               1.10.0
   ```

   If the command above hangs, you might need to adjust your firewall rules.
   Your old Pachyderm cluster can operate while you are creating a migrated
   one.

1. Proceed to [Step 3](#step-3-deploy-a-pachyderm-cluster-with-the-cloned-bucket).

## Step 3 - Deploy a Pachyderm Cluster with the Cloned Bucket

After you create a backup of your existing cluster, you need to create a new
Pachyderm cluster by using the bucket you cloned in [Step 1](#step-1-back-up-your-cluster).

This new cluster can be deployed:

* On the same Kubernetes cluster in a separate namespace.
* On a different Kubernetes cluster within the same cloud provider.

If you are deploying in a namespace on the same Kubernetes cluster,
you might need to modify Kubernetes ingress to Pachyderm deployment in the
new namespace to avoid port conflicts in the same cluster.
Consult with your Kubernetes administrator for information on avoiding
ingress conflicts.

If you have issues with the extracted data, rerun instructions in
[Step 1](#step-1-back-up-your-cluster).

To deploy a Pachyderm cluster with a cloned bucket, complete the following
steps:

1. Upgrade your Pachyderm version to the latest version:

   ```shell
   brew upgrade pachyderm/tap/pachctl@{{config.pach_major_minor_version}}
   ```

   * If you are deploying your cluster in a separate Kubernetes namespace,
   create a new namespace:

  ```shell
  kubectl create namespace <new-cluster-namespace>
  ```

1. Deploy your cluster in a separate namespace or on a separate Kubernetes
cluster by using a `pachctl deploy` command for your cloud provider with the
`--namespace` flag.

   **Examples:**

   - AWS EKS
      ```shell
      pachctl deploy amazon <bucket-name> <region> <storage-size> --dynamic-etcd-nodes=<number> --iam-role <iam-role> --namespace=<namespace-name>
      ```

   - GKE
      ```shell
      pachctl deploy google <bucket-name> <storage-size> --dynamic-etcd-nodes=1  --namespace=<namespace-name>
      ```

   - Azure
      ```shell
      pachctl deploy microsoft <account-name> <storage-account> <storage-key> <storage-size> --dynamic-etcd-nodes=<number> --namespace=<namespace-name>
      ```

   **Note:** Parameters for your Pachyderm cluster deployment might be different.
   For more information, see [Deploy Pachyderm](../../deploy/).

1. Verify that your cluster has been deployed:

- In a namespace
    ```shell
    kubectl get pod --namespace=<new-cluster>
    ```

- On a cluster
    ```shell
    kubectl get pod
    ```

   If you have deployed your new cluster in a namespace, Pachyderm should
   have created a new context for this deployement. Verify that you are
   using this.

1. Proceed to [Step 4](#step-4-restore-your-cluster).

## Step 4 - Restore your Cluster

After you have created a new cluster, you can restore your backup to this
new cluster. If you have deployed your new cluster in a namespace, Pachyderm
should have created a new context for this deployement. You need to switch to
this new context to access the correct cluster. Before you run the
`pachctl restore` command, your new cluster should be empty.

To restore your cluster, complete the following steps:

1. If you deployed your new cluster into a different namespace on the same
   Kubernetes cluster as your old cluster, verify that you on the correct namespace:

  ```shell
  pachctl config get context `pachctl config get active-context`
  ```

  **Example System Response:**

  ``` json
  {
    "source": "IMPORTED",
    "cluster_name": "test-migration.us-east-1.eksctl.io",
    "auth_info": "user@test-migration.us-east-1.eksctl.io",
    "namespace": "new-cluster"
  }
  ```

  Your active context must have the namespace you have deployed your new
  cluster into.

1. Check that the cluster does not have any exisiting Pachyderm objects:

   ```shell
   pachctl list repo 
   pachctl list pipeline
   ```

   You should get empty output.

1. Restore your cluster from the backup you have created in
[Step 1](#step-1-back-up-your-cluster):

   - Local File
      ```shell
      pachctl restore < path/to/your/backup/file
      ```

   - S3 Bucket
      ```shell
      pachctl restore --url s3://path/to/backup
      ```


   This S3 bucket is different from the s3 bucket to which you cloned
   your Pachyderm data. This is merely a bucket you allocated to hold
   the Pachyderm backup without objects.

1. Configure any external data loading systems to point at the new,
upgraded Pachyderm cluster and play back transactions from the checkpoint
established at [Pause External Data Operations](./backup-migrations/#pause-external-data-loading-operations).
Perform any reconfiguration to data loading or unloading operations.
Confirm that the data output is as expected and the new cluster is operating as expected.

1. Disable the old cluster:

   * If you have deployed the new cluster on the same Kuberenetes cluster
   switch to the old cluster's Pachyderm context:

      ```shell
      pachctl config set active-context <old-context>
      ```

   * If you have deployed the new cluster to a different Kubernetes cluster,
   switch to the old cluster's Kubernetes context:

      ```shell
      kubectl config use-context <old cluster>
      ```

1. Undeploy your old cluster:

      ```shell
      pachctl undeploy
      ```

1. Reconfigure new cluster as necessary
   You may need to reconfigure the following:

   - Data loading operations from Pachyderm to processes outside
   of it to work as expected.
   - Kubernetes ingress and port changes taken to avoid conflicts
   with the old cluster.
