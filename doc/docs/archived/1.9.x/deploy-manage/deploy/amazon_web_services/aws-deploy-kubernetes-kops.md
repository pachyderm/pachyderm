# Deploy Kubernetes with `kops`

`kops` is one of the most popular open-source tools
that enable you to deploy, manage, and upgrade a
Kubernetes cluster in the cloud. By using `kops` you can
quickly spin-up a highly-available Kubernetes cluster in
a supported cloud platform.

## Prerequisites

Before you can deploy Pachyderm on Amazon AWS with
`kops`, you must have the following components configured:

- Install [AWS CLI](https://aws.amazon.com/cli/)
- Install [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- Install [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)
- Install [pachctl](../../../../getting_started/local_installation/#install-pachctl)
- Install [jq](https://stedolan.github.io/jq/download/)
- Install [uuid](http://man7.org/linux/man-pages/man1/uuidgen.1.html)

## Configure `kops`

[`kops`](https://github.com/kubernetes/kops/), which stands for
*Kubernetes Operations*, is an open-source tool that deploys
a production-grade Kubernetes cluster on a cloud environment of choice.
You need to have access to the
AWS Management console to add an Identity and Access Management (IAM) user
for `kops`.

For more information about `kops`, see
[kops AWS documentation](https://github.com/kubernetes/kops/blob/master/docs/aws.md).
These instructions provide more details about configuring
additional cluster parameters, such as enabling version control
or encryption on your S3 bucket, and so on.

To configure `kops`, complete the following steps:

1. In the IAM console or by using the command line, create a `kops` group
with the following permissions:

   * `AmazonEC2FullAccess`
   * `AmazonRoute53FullAccess`
   * `AmazonS3FullAccess`
   * `IAMFullAccess`
   * `AmazonVPCFullAccess`

1. Add a user that will create a Kubernetes cluster to that group.
1. In the list of users, select that user and navigate to the
**Security credentials** tab.
1. Create an access key and save the access and secret keys in a
location on your computer.
1. Configure an AWS CLI client:

   ```shell
   $ aws configure
   ```

1. Use the access and secret keys to configure the AWSL client.

1. Create an S3 bucket for your cluster:

   ```shell
   $ aws s3api create-bucket --bucket <name> --region <region>
   ```

   **Example:**

   ```shell
   $ aws s3api create-bucket --bucket test-pachyderm --region us-east-1
   {
        "Location": "/test-pachyderm"
   }
   ```

1. Optionally, configure DNS as described in [Configure DNS](https://github.com/kubernetes/kops/blob/master/docs/aws.md#configure-dns).
In this example, a gossip-based cluster that ends with `k8s.local`
is deployed.

1. Export the name of your cluster and the S3 bucket for the Kubernetes
cluster as variables.

   **Example:**

   ```shell
   export NAME=test-pachyderm.k8s.local
   export KOPS_STATE_STORE=s3://test-pachyderm
   ```

1. Create the cluster configuration:

   ```shell
   kops create cluster --zones <region> ${NAME}
   ```

1. Optionally, edit your cluster:

   ```shell
   kops edit cluster ${NAME}
   ```

1. Build and deploy the cluster:

   ```shell
   kops update cluster ${NAME} --yes
   ```

   The deployment might take some time.

1. Run `kops cluster validate` periodically to monitor cluster deployment.
   When `kops` finishes deploying the cluster, you should see the output
   similar to the following:

   ```shell
   $ kops validate cluster
   Using cluster from kubectl context: test-pachyderm.k8s.local

   Validating cluster svetkars.k8s.local

   INSTANCE          GROUPS
   NAME              ROLE     MACHINETYPE MIN MAX SUBNETS
   master-us-west-2a Master   m3.medium   1   1   us-west-2a
   nodes             Node     t2.medium   2   2   us-west-2a

   NODE                                                   STATUS
   NAME                                           ROLE    READY
   ip-172-20-45-231.us-west-2.compute.internal    node    True
   ip-172-20-50-8.us-west-2.compute.internal      master  True
   ip-172-20-58-132.us-west-2.compute.internal    node    True
   ```

1. Proceed to [Deploy Pachyderm on AWS](aws-deploy-pachyderm.md).
