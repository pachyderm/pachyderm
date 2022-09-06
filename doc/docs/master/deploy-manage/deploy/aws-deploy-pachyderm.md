# Deploy Pachyderm on AWS

For a quick test installation of Pachyderm on AWS (suitable for development), jump to our [Quickstart page](../quickstart/).

For deployments in production, refer to the following diagram and follow these step-by-step instructions:
![AWS Arch](../images/arch-diagram-high-level-aws.svg)

!!! Important "Before your start your installation process." 
      - Refer to our generic ["Helm Install"](../helm-install/){target=_blank} page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/){target=_blank}. You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
      - Pachyderm comes with a [web UI (Console)](../console) for visualizing running pipelines and exploring your data. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up of an Ingress.
    

!!! Attention 
    We are now shipping Pachyderm with an **embedded proxy** 
    allowing your cluster to expose one single port externally. This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a Proxy, check out our new recommended architecture and [deployment instructions](../deploy-w-proxy/) as they alter the instructions below.

The following section walks you through deploying a Pachyderm cluster on [Amazon Elastic Kubernetes Service](https://aws.amazon.com/eks/){target=_blank} (EKS). 

In particular, you will:

1. Make a few [client installations](#1-prerequisites) before you start.
1. [Deploy Kubernetes](#2-deploy-kubernetes-by-using-eksctl).
1. [Create an S3 bucket](#3-create-an-s3-bucket) for your data and grant Pachyderm access.
1. [Enable Persistent Volumes Creation](#4-enable-your-persistent-volumes-creation)
1. [Create An AWS Managed PostgreSQL Instance](#5-create-an-aws-managed-postgresql-database)
1. [Deploy Pachyderm ](#6-deploy-pachyderm)
1. Finally, you will need to install [pachctl](../../../getting-started/local-installation#install-pachctl) to [interact with your cluster](#7-have-pachctl-and-your-cluster-communicate).
1. And check that your cluster is [up and running](#8-check-that-your-cluster-is-up-and-running)
1. (Optional) Install [JupyterHub and Pachyderm Mount Extension](#9-notebooks-users-install-pachyderm-jupyterlab-mount-extension) to experiment with your data in Pachyderm from your Notebook cells. 

## 1. Prerequisites

Before you can deploy Pachyderm on an EKS cluster, verify that
you have the following prerequisites installed and configured:

* [kubectl](https://kubernetes.io/docs/tasks/tools/){target=_blank}
* [AWS CLI](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html){target=_blank}
* [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html){target=_blank}
* [aws-iam-authenticator](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html){target=_blank}.
* [pachctl](../../../getting-started/local-installation#install-pachctl){target=_blank}

## 2. Deploy Kubernetes by using `eksctl`

!!! Attention
      Pachyderm requires running your cluster on Kubernetes 1.19.0 and above.

Use the `eksctl` tool to deploy an EKS cluster in your
Amazon AWS environment. The `eksctl create cluster` command
creates a virtual private cloud (VPC), a security group,
and an IAM role for Kubernetes to create resources.
For detailed instructions, see [Amazon documentation](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html){target=_blank}.

To deploy an EKS cluster, complete the following steps:

1. Deploy an EKS cluster:

      ```shell
      eksctl create cluster --name <name> --version <version> \
      --nodegroup-name <name> --node-type <vm-flavor> \
      --nodes <number-of-nodes> --nodes-min <min-number-nodes> \
      --nodes-max <max-number-nodes> --node-ami auto
      ```

      **Example**

      ```shell
      eksctl create cluster --name pachyderm-cluster --region us-east-2 --profile <your named profile>
      ```

1. Verify the deployment:

      ```shell
      kubectl get all
      ```

      **System Response:**

      ```
      NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
      service/kubernetes   ClusterIP   10.100.0.1   <none>        443/TCP   23h
      ```

Once your Kubernetes cluster is up, and your infrastructure is configured, 
you are ready to prepare for the installation of Pachyderm.
Some of the steps below will require you to keep updating the values.yaml started during the setup of the recommended infrastructure. 

!!! Note "Secrets Manager"
      Pachyderm recommends securing and managing your secrets in a Secret Manager. Learn about the [set up and configuration of your EKS cluster to retrieve the relevant secrets from AWS Secrets Manager](../aws-secret-manager){target=_blank} then resume the following installation steps.

## 3. Create an S3 bucket
### Create an S3 object store bucket for data

Pachyderm needs an S3 bucket (Object store) to store your data. You can create the bucket by running the following commands:

!!! Warning
      The S3 bucket name must be globally unique across the entire
      Amazon region. 

* Set up the following system variables:

      * `BUCKET_NAME` — A globally unique S3 bucket name.
      * `AWS_REGION` — The AWS region of your Kubernetes cluster. For example,
      `us-west-2` and not `us-west-2a`.

* If you are creating an S3 bucket in the `us-east-1` region, run the following
      command:

      ```shell
      aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}
      ```

* If you are creating an S3 bucket in any region but the `us-east-1`
region, run the following command:

      ```shell
      aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
      ```

* Verify that the S3 bucket was created:

      ```shell   
      aws s3 ls
      ```

You now need to **give Pachyderm access to your bucket** either by:

- [Adding a policy to your service account IAM Role](#add-an-iam-role-and-policy-to-your-service-account) (Recommended)
OR
- Passing your AWS credentials (account ID and KEY) to your values.yaml when installing

!!! Info
      IAM roles provide finer grained user management and security
      capabilities than access keys. Pachyderm recommends the use of IAM roles for production
      deployments.

### Add An IAM Role And Policy To Your Service Account

Before you can make sure that **the containers in your pods have the right permissions to access your S3 bucket**, you will need to [Create an IAM OIDC provider for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html){target=_blank}.

Then follow the steps detailled in **[Create an IAM Role And Policy for your Service Account](https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html){target=_blank}**.

In short, you will:

1. Retrieve your **OpenID Connect provider URL**:
      1. Go to the AWS Management console.
      1. Select your cluster instance in **Amazon EKS**.
      1. In the **Configuration** tab of your EKS cluster, find your **OpenID Connect provider URL** and save it. You will need it when creating your IAM Role.

1. Create an **IAM policy** that gives access to your bucket:
      1. Create a new **Policy** from your IAM Console.
      1. Select the **JSON** tab.
      1. Copy/Paste the following text in the JSON tab:

      ```json
      {
            "Version": "2012-10-17",
            "Statement": [
                  {
            "Effect": "Allow",
                  "Action": [
                        "s3:ListBucket"
                  ],
                  "Resource": [
                        "arn:aws:s3:::<your-bucket>"
                  ]},{
            "Effect": "Allow",
                  "Action": [
                        "s3:PutObject",
                        "s3:GetObject",
                        "s3:DeleteObject"
                  ],
                  "Resource": [
                        "arn:aws:s3:::<your-bucket>/*"
                  ]}
            ]
      }
      ``` 

      Replace `<your-bucket>` with the name of your S3 bucket.

1. Create an **IAM role as a Web Identity** using the cluster OIDC procider as the identity provider.
      1. Create a new **Role** from your IAM Console.
      1. Select the **Web identity** Tab.
      1. In the **Identity Provider** drop down, select the *OpenID Connect provider URL* of your EKS and `sts.amazonaws.com` as the Audience.
      1. Attach the newly created permission to the Role.
      1. Name it.
      1. Retrieve the **Role arn**. You will need it in your values.yaml annotations when deploying Pachyderm.

### (Optional) Set Up Bucket Encryption

Amazon S3 supports two types of bucket encryption — server-side encryption
(SSE-S3) and AWS Key Management Service (AWS KMS), which stores customer
master keys. When creating a bucket for your Pachyderm cluster, you can set up either
of them. Because Pachyderm requests that buckets do not include encryption
information, the method that you select for the bucket is applied.

!!! Info
      Setting up communication between Pachyderm object storage clients and AWS KMS
      to append encryption information to Pachyderm requests is not supported and
      not recommended. 

To set up bucket encryption, see [Amazon S3 Default Encryption for S3 Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html){target=_blank}.

## 4. Enable Your Persistent Volumes Creation

etcd and PostgreSQL (metadata storage) each claim the creation of a pv. 

!!! Important
      The metadata services generally require a small persistent volume size (i.e. 10GB) **but high IOPS (1500)**.
      Note that Pachyderm out-of-the-box deployment comes with **gp2** default EBS volumes. 
      While it might be easier to set up for test or development environments, **we highly recommend to use SSD gp3 in production**. A **gp3** EBS volume delivers a baseline performance of 3,000 IOPS and 125MB/s at any volume size. Any other disk choice may require to **oversize the volume significantly to ensure enough IOPS**.

      See [volume types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html){target=_blank}.

If you plan on using **gp2** EBS volumes:

- For deployments in **production**, [go to the AWS-managed PostgreSQL deployment and setup section](#5-create-an-aws-managed-postgresql-database)
- For non production deployments, you will be using the default bundled version of PostgreSQL: [Go to the deployment of Pachyderm](#6-deploy-pachyderm) 

For gp3 volumes, you will need to **deploy an Amazon EBS CSI driver to your cluster as detailed below**.

For your EKS cluster to successfully create two **Elastic Block Storage (EBS) persistent volumes (PV)**, follow the steps detailled in **[deploy Amazon EBS CSI driver to your cluster](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html){target=_blank}**.

In short, you will:

1. [Create an IAM OIDC provider for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html){target=_blank}. You might already have completed this step if you chose to create an IAM Role and Policy to give your containers permission to access your S3 bucket.
1. Create a CSI Driver service account whose IAM Role will be granted the permission (policy) to make calls to AWS APIs. 
1. Install Amazon EBS Container Storage Interface (CSI) driver on your cluster configured with your created service account.


If you expect your cluster to be very long
running or scale to thousands of jobs per commits, you might need to add
more storage.  However, you can easily increase the size of the persistent
volume later.

## 5. Create an AWS Managed PostgreSQL Database

By default, Pachyderm runs with a bundled version of PostgreSQL. 
For production environments, it is **strongly recommended that you disable the bundled version and use an RDS PostgreSQL instance**. 

!!! Attention
      Note that [Aurora Serverless PostgreSQL](https://aws.amazon.com/rds/aurora/serverless/){target=_blank} is not supported and will not work.

This section will provide guidance on the configuration settings you will need to: 

- Create an environment to run your AWS PostgreSQL databases. Note that you might need to 
create **two databases** (`pachyderm` and, depending on whether your cluster is standalone or managed by an enterprise server, a second database, `dex`).
- Update your values.yaml to turn off the installation of the bundled postgreSQL and provide your new instance information.

!!! Note
      It is assumed that you are already familiar with RDS, or will be working with an administrator who is.

### Create An RDS Instance

!!! Info 
      Find the details of all the steps highlighted below in [AWS Documentation: "Getting Started" hands-on tutorial](https://aws.amazon.com/getting-started/hands-on/create-connect-postgresql-db/){target=_blank}.
 
In the RDS console, create a database **in the region matching your Pachyderm cluster**. Choose the **PostgreSQL** engine and select a PostgreSQL version >= 13.3.

Configure your DB instance as follows.

| SETTING | Recommended value|
|:----------------|:--------------------------------------------------------|
| *DB instance identifier* | Fill in with a unique name across all of your DB instances in the current region.|
| *Master username* | Choose your Admin username.|
| *Master password* | Choose your Admin password.|
| *DB instance class* | The standard default should work. You can change the instance type later on to optimize your performances and costs. |
| *Storage type* and *Allocated storage*| If you choose **gp2**, remember that Pachyderm's metadata services require **high IOPS (1500)**. Oversize the disk accordingly (>= 1TB). <br> If you select **io1**, keep the 100 GiB default size. <br> Read more [information on Storage for RDS on Amazon's website](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html){target=_blank}. |
| *Storage autoscaling* | If your workload is cyclical or unpredictable, enable storage autoscaling to allow RDS to scale up your storage when needed. |
| *Standby instance* | We highly recommend creating a standby instance for production environments.|
| *VPC* | **Select the VPC of your Kubernetes cluster**. Attention: After a database is created, you can't change its VPC. <br> Read more on [VPCs and RDS on Amazon documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_VPC.html){target=_blank}.| 
| *Subnet group* | Pick a Subnet group or Create a new one. <br> Read more about [DB Subnet Groups on Amazon documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_VPC.WorkingWithRDSInstanceinaVPC.html#USER_VPC.Subnets){target=_blank}. |
| *Public access* | Set the Public access to `No` for production environments. |
| *VPC security group* | Create a new VPC security group and open the postgreSQL port or use an existing one. |
| *Password authentication* or *Password and IAM database authentication* | Choose one or the other. |
| *Database name* | In the *Database options* section, enter Pachyderm's Database name (We are using `pachyderm` in this example.) and click *Create database* to create your PostgreSQL service. Your instance is running. <br>Warning: If you do not specify a database name, Amazon RDS does not create a database.|


!!! Warning "One last step"
      Once your instance is created: 

      - If you plan to deploy a standalone cluster (i.e., if you do not plan to register your cluster with a separate [enterprise server](../../enterprise/auth/enterprise-server/setup.md), you will need to create a second database named "dex" in your RDS instance for Pachyderm's authentication service. Note that the database **must be named `dex`**. Read more about [dex on PostgreSQL in Dex's documentation](https://dexidp.io/docs/storage/#postgres){target=_blank}. This second database is not needed when your cluster is managed by an enterprise server.
      - Additionally, create a new user account and **grant it full CRUD permissions to both `pachyderm` and (when applicable) `dex` databases**. Read about managing PostgreSQL users and roles in this [blog](https://aws.amazon.com/blogs/database/managing-postgresql-users-and-roles/). Pachyderm will use the same username to connect to `pachyderm` as well as to `dex`. 

### Update your values.yaml 
Once your databases have been created, add the following fields to your Helm values:

```yaml
global:
  postgresql:
    postgresqlUsername: "username"
    postgresqlPassword: "password" 
    # The name of the database should be Pachyderm's ("pachyderm" in the example above), not "dex" 
    # See also 
    # postgresqlExistingSecretName: "<yoursecretname>"
    postgresqlDatabase: "databasename"
    # The postgresql database host to connect to. Defaults to postgres service in subchart
    postgresqlHost: "RDS CNAME"
    # The postgresql database port to connect to. Defaults to postgres server in subchart
    postgresqlPort: "5432"

postgresql:
  # turns off the install of the bundled postgres.
  # If not using the built in Postgres, you must specify a Postgresql
  # database server to connect to in global.postgresql
  enabled: false
```
## 6. Deploy Pachyderm

You have set up your infrastructure, created your S3 bucket and an AWS Managed PostgreSQL instance, and granted your cluster access to both: you can now finalize your values.yaml and deploy Pachyderm.

### Update Your Values.yaml  

!!! Note
     If you have not created a Managed PostgreSQL RDS instance, **replace the Postgresql section below** with `postgresql:enabled: true` in your values.yaml. This setup is **not recommended in production environments**.
#### For gp3 EBS Volumes

[Check out our example of values.yaml for gp3](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/aws-gp3-values.yaml){target=_blank} or use our minimal example below.


=== "Gp3 + Service account annotations"   
      ```yaml
      deployTarget: AMAZON
      # This uses GP3 which requires the CSI Driver https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html
      # And a storageclass configured named gp3
      etcd:
        storageClass: gp3
      pachd:
        storage:
          amazon:
            bucket: blah
            region: us-east-2
        serviceAccount:
          additionalAnnotations:
            eks.amazonaws.com/role-arn: arn:aws:iam::<ACCOUNT_ID>:role/pachyderm-bucket-access
        worker:
          serviceAccount:
            additionalAnnotations:
              eks.amazonaws.com/role-arn: arn:aws:iam::<ACCOUNT_ID>:role/pachyderm-bucket-access
        externalService:
          enabled: true
      global:
        postgresql:
          postgresqlUsername: "username"
          postgresqlPassword: "password" 
          # The name of the database should be Pachyderm's ("pachyderm" in the example above), not "dex" 
          postgresqlDatabase: "databasename"
          # The postgresql database host to connect to. Defaults to postgres service in subchart
          postgresqlHost: "RDS CNAME"
          # The postgresql database port to connect to. Defaults to postgres server in subchart
          postgresqlPort: "5432"

      postgresql:
        # turns off the install of the bundled postgres.
        # If not using the built in Postgres, you must specify a Postgresql
        # database server to connect to in global.postgresql
        enabled: false
      ```
=== "Gp3 + AWS Credentials"   
      ```yaml
      deployTarget: AMAZON
      # This uses GP3 which requires the CSI Driver https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html
      # And a storageclass configured named gp3
      etcd:
        storageClass: gp3
      pachd:
        storage:
          amazon:
            bucket: blah
            region: us-east-2
            # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            id: AKIAIOSFODNN7EXAMPLE
            # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
        externalService:
          enabled: true           
      global:
        postgresql:
          postgresqlUsername: "username"
          postgresqlPassword: "password" 
          # The name of the database should be Pachyderm's ("pachyderm" in the example above), not "dex" 
          postgresqlDatabase: "databasename"
          # The postgresql database host to connect to. Defaults to postgres service in subchart
          postgresqlHost: "RDS CNAME"
          # The postgresql database port to connect to. Defaults to postgres server in subchart
          postgresqlPort: "5432"

      postgresql:
        # turns off the install of the bundled postgres.
        # If not using the built in Postgres, you must specify a Postgresql
        # database server to connect to in global.postgresql
        enabled: false
      ```

#### For gp2 EBS Volumes

[Check out our example of values.yaml for gp2](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/aws-gp2-values.yaml){target=_blank} or use our minimal example below.   
    
=== "For Gp2 + Service account annotations"
      ```yaml
      deployTarget: AMAZON      
      etcd:
        etcd.storageSize: 500Gi
      pachd:
        storage:
          amazon:
            bucket: blah
            region: us-east-2
        serviceAccount:
          additionalAnnotations:
            eks.amazonaws.com/role-arn: arn:aws:iam::190146978412:role/pachyderm-bucket-access
        worker:
          serviceAccount:
            additionalAnnotations:
              eks.amazonaws.com/role-arn: arn:aws:iam::190146978412:role/pachyderm-bucket-access
        externalService:
          enabled: true
      global:
        postgresql:
          postgresqlUsername: "username"
          postgresqlPassword: "password" 
          # The name of the database should be Pachyderm's ("pachyderm" in the example above), not "dex" 
          postgresqlDatabase: "databasename"
          # The postgresql database host to connect to. Defaults to postgres service in subchart
          postgresqlHost: "RDS CNAME"
          # The postgresql database port to connect to. Defaults to postgres server in subchart
          postgresqlPort: "5432"

      postgresql:
        # turns off the install of the bundled postgres.
        # If not using the built in Postgres, you must specify a Postgresql
        # database server to connect to in global.postgresql
        enabled: false
      ```  
=== "For Gp2 + AWS Credentials"
      ```yaml
      deployTarget: AMAZON      
      etcd:
        etcd.storageSize: 500Gi
      pachd:
        storage:
          amazon:
            bucket: blah
            region: us-east-2
            # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            id: AKIAIOSFODNN7EXAMPLE            
            # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html           
            secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
        externalService:
          enabled: true
      global:
        postgresql:
          postgresqlUsername: "username"
          postgresqlPassword: "password" 
          # The name of the database should be Pachyderm's ("pachyderm" in the example above), not "dex" 
          postgresqlDatabase: "databasename"
          # The postgresql database host to connect to. Defaults to postgres service in subchart
          postgresqlHost: "RDS CNAME"
          # The postgresql database port to connect to. Defaults to postgres server in subchart
          postgresqlPort: "5432"

      postgresql:
        # turns off the install of the bundled postgres.
        # If not using the built in Postgres, you must specify a Postgresql
        # database server to connect to in global.postgresql
        enabled: false
      ```


Check the [list of all available helm values](../../../reference/helm-values/) at your disposal in our reference documentation or on [Github](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/pachyderm/values.yaml){target=_blank}.

!!! Important
      Retain (ideally in version control) a copy of the Helm values used to deploy your cluster. It might be useful if you need to [restore a cluster from a backup](../../manage/backup-restore).
### Deploy Pachyderm On The Kubernetes Cluster


- You can now deploy a Pachyderm cluster by running this command:

      ```shell
      helm repo add pach https://helm.pachyderm.com
      helm repo update
      helm install pachyderm -f values.yaml pach/pachyderm --version <version-of-the-chart>
      ```

      **System Response:**

      ```shell
      NAME: pachd
      LAST DEPLOYED: Mon Jul 12 18:28:59 2021
      NAMESPACE: default
      STATUS: deployed
      REVISION: 1
      ```

      The deployment takes some time. You can run `kubectl get pods` periodically
      to check the status of deployment. When Pachyderm is deployed, the command
      shows all pods as `READY`:

      ```shell
      kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m
      ```

      **System Response**

      ```
      pod/pachd-74c5766c4d-ctj82 condition met
      ```

      **Note:** If you see a few restarts on the `pachd` nodes, it means that
      Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
      Kubernetes restarted those pods. You can safely ignore this message.

- Finally, make sure that [`pachctl` talks with your cluster](#7-have-pachctl-and-your-cluster-communicate).

## 7. Have 'pachctl' And Your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you are exposing your cluster publicly:

  1. Retrieve the external IP address of your TCP load balancer or your domain name:
    ```shell
    kubectl get services | grep pachd-lb | awk '{print $4}'
    ```

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
      ```
      ```shell
      pachctl config set active-context "<your-cluster-context-name>"
      ```

  1. Check that your are using the right context: 

      ```shell
      pachctl config get active-context
      ```

      Your cluster context name should show up.

If you're not exposing `pachd` publicly, you can run:

```shell
# Background this process because it blocks.
pachctl port-forward
``` 

## 8. Check That Your Cluster Is Up And Running

!!! Attention
    If Authentication is activated (When you deploy with an enterprise key already set, for example), you need to run `pachct auth login`, then authenticate to Pachyderm with your User, before you use `pachctl`. 

```shell
pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

## 9. NOTEBOOKS USERS: Install Pachyderm JupyterLab Mount Extension

Once your cluster is up and running, you can helm install JupyterHub on your Pachyderm cluster and experiment with your data in Pachyderm from your Notebook cells. 

Check out our [JupyterHub and Pachyderm Mount Extension](../../how-tos/jupyterlab-extension/index.md){target=_blank} page for installation instructions. 

Use Pachyderm's default image and values.yaml [`jupyterhub-ext-values.yaml`](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/jupyterhub-ext-values.yaml){target=_blank} or follow the instructions to update your own.

!!! Note
       Make sure to check our [data science notebook examples](https://github.com/pachyderm/examples){target=_blank} running on Pachyderm, from a market sentiment NLP implementation using a FinBERT model to pipelines training a regression model on the Boston Housing Dataset.