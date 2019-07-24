# Deploy by Executing a Deployment Script

This section describes how to deploy a Kubernetes and Pachyderm
cluster by using a single deployment
script. This deployment is designed for testing environments.
The deployments script uses `kops` to deploy Kubernetes on
AWS. `kops` which stands for *Kubernetes Operations*, is a tool that
deploys a production-grade Kubernetes cluster on your cloud environment
of choice.

## Prerequisites

Before you can deploy Pachyderm by using the AWS deployment script,
verify that you have configured the following prerequisites:

- Install [AWS CLI](https://aws.amazon.com/cli/)
- Configure [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)
- Install [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- Install [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)
- Install [pachctl](#install-pachctl)
- Install [jq](https://stedolan.github.io/jq/download/)
- Install [uuid](http://man7.org/linux/man-pages/man1/uuidgen.1.html)

## Run the deployment script

After you have configured the prerequisites mentioned above,
download and run the AWS deployment script. This script uses `kops`
to deploy Kubernetes and Pachyderm in Amazon Web Services (AWS).
The script prompts you to specify your AWS credentials, region
preference, and so on. If you want to customize the number of
nodes in the cluster, node types, and other parameters,
you can open the AWS script source file and modify the respective
fields directly in the file.

To deploy Pachyderm by using the deployment script,
complete the following steps:

1. Fetch the deployment script from the Pachyderm repository:

   ```
   $ curl -o aws.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/aws.sh
   ```

1. Make the deployment script executable:

   ```
   $ chmod +x aws.sh
   ```
1. Run the deployment script:

   ```
   $ sudo -E ./aws.sh
   ```

   It might take a few minutes for the script to complete.
   You can run `kubectl get pods`
   periodically to check the deployment progress. When the
   deployment is complete, `kubectl get pods` shows that
   all pods are ready.

   ```bash
   $ kubectl get pods
   NAME                     READY     STATUS    RESTARTS   AGE
   dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
   etcd-0                   1/1       Running   0          4m
   pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
   ```

## Access your Pachyderm Cluster

To access your Pachyderm cluster and execute `pachctl` commands,
enable port forwarding by completing the following steps:

1. Open a new terminal window:

   ```bash
   $ pachctl port-forward &
   ```

   This command runs continuously and does not exit, unless
   you interrupt it.

1. Open a separate terminal window and verify that you can run
   `pachctl` commands:

   ```bash
   $ pachctl version
   COMPONENT           VERSION
   pachctl             1.9.1
   pachd               1.9.1
   ```

## Remove a Pachyderm Cluster

You can delete your Pachyderm cluster by using the `kops delete cluster`
command. Also, you need to manually remove an entry in the `/etc/hosts`
file and any S3 buckets that are associated with the cluster.

To remove a Pachyderm cluster, complete the following steps:

1. Delete the Pachyderm cluster:

   ```bash
   $ kops delete cluster
   ```

1. Open the `/etc/hosts` file and delete the entry that points to the
cluster.

1. In the AWS Management Console, delete the Kubernetes
S3 bucket and Pachyderm storage bucket.
