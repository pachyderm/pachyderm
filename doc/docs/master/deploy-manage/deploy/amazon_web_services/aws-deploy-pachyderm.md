# Deploy Pachyderm on AWS

!!! Important "Before your start your installation process." 
      - Refer to our generic ["Helm Install"](./helm_install.md) page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../../ingress/). You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
      - If you are planning to install Pachyderm UI. Read our [Console deployment](../../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up on an [Ingress](../../ingress/#ingress).

Once your Kubernetes cluster is up, and your infrastructure is configured, 
you are ready to prepare for the installation of Pachyderm. 
Some of the steps below will require you to keep updating the values.yaml started during the setup of the recommended infrastructure:


1. [Create an S3 bucket](#1-create-an-S3-object) for your data and grant Pachyderm access.
1. [Enable Persistent Volumes Creation](#2-enable-your-persistent-volumes-creation)
1. [Create An AWS Managed PostgreSQL Instance](#3-create-an-aws-managed-postgresql-database)
1. [Deploy Pachyderm ](#4-deploy-pachyderm)
1. Finally, you will need to install [pachctl](../../../../getting_started/local_installation#install-pachctl) to [interact with your cluster](#5-have-pachctl-and-your-cluster-communicate).
1. And check that your cluster is [up and running](#6-check-that-your-cluster-is-up-and-running)

## 1. Create an S3 bucket
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
      $ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}
      ```

* If you are creating an S3 bucket in any region but the `us-east-1`
region, run the following command:

      ```shell
      $ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
      ```

* Verify that the S3 bucket was created:

      ```shell   
      $ aws s3 ls
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

Before you can make sure that **the containers in your pods have the right permissions to access your S3 bucket**, you will need to [Create an IAM OIDC provider for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

Then follow the steps detailled in **[Create an IAM Role And Policy for your Service Account](https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html)**.

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

To set up bucket encryption, see [Amazon S3 Default Encryption for S3 Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html).

## 2. Enable Your Persistent Volumes Creation

etcd and PostgreSQL (metadata storage) each claim the creation of a pv. 

!!! Important
      The metadata services generally require a small persistent volume size (i.e. 10GB) **but high IOPS (1500)**.
      Note that Pachyderm out-of-the-box deployment comes with **gp2** default EBS volumes. 
      While it might be easier to set up for test or development environments, **we highly recommend to use SSD gp3 in production**. A **gp3** EBS volume delivers a baseline performance of 3,000 IOPS and 125MB/s at any volume size. Any other disk choice may require to **oversize the volume significantly to ensure enough IOPS**.

      See [volume types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html).

If you plan on using **gp2** EBS volumes:

- [Skip this section and jump to the deployment of Pachyderm](#3-deploy-pachyderm) 
- or, for deployments in production, [jump to AWS-managed PostgreSQL](#3-optional-amazon-rds-aws-managed-postgresql-database)

For gp3 volumes, you will need to **deploy an Amazon EBS CSI driver to your cluster as detailed below**.

For your EKS cluster to successfully create two **Elastic Block Storage (EBS) persistent volumes (PV)**, follow the steps detailled in **[deploy Amazon EBS CSI driver to your cluster](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html)**.

In short, you will:

1. [Create an IAM OIDC provider for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html). You might already have completed this step if you choose to create an IAM Role and Policy to give your containers permission to access your S3 bucket.
1. Create a CSI Driver service account whose IAM Role will be granted the permission (policy) to make calls to AWS APIs. 
1. Install Amazon EBS Container Storage Interface (CSI) driver on your cluster configured with your created service account.


If you expect your cluster to be very long
running or scale to thousands of jobs per commits, you might need to add
more storage.  However, you can easily increase the size of the persistent
volume later.

## 3. Create an AWS Managed PostgreSQL Database

By default, Pachyderm runs with a bundled version of PostgreSQL. 
For production environments, it is **strongly recommended that you disable the bundled version and use an RDS PostgreSQL instance**. 

This section will provide guidance on the configuration settings you will need to: 

- Create an environment to run your AWS PostgreSQL databases. Note that you will be creating **two databases** (`pachyderm` and `dex`).
- Update your values.yaml to turn off the installation of the bundled postgreSQL and provide your new instance information.

!!! Note
      It is assumed that you are already familiar with RDS, or will be working with an administrator who is.

### Create An RDS Instance

!!! Info 
      Find the details of all the steps highlighted below in [AWS Documentation: "Getting Started" hands-on tutorial](https://aws.amazon.com/getting-started/hands-on/create-connect-postgresql-db/).
 
In the RDS console, create a database **in the region matching your Pachyderm cluster**. Choose the **PostgreSQL** engine and select a PostgreSQL version >= 13.3.

Configure your DB instance as follows.

| SETTING | Recommended value|
|:----------------|:--------------------------------------------------------|
| *DB instance identifier* | Fill in with a unique name across all of your DB instances in the current region.|
| *Master username* | Choose your Admin username.|
| *Master password* | Choose your Admin password.|
| *DB instance class* | The standard default should work. You can change the instance type later on to optimize your performances and costs. |
| *Storage type* and *Allocated storage*| If you choose **gp2**, remember that Pachyderm's metadata services require **high IOPS (1500)**. Oversize the disk accordingly (>= 1TB). <br> If you select **io1**, keep the 100 GiB default size. <br> Read more [information on Storage for RDS on Amazon's website](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html). |
| *Storage autoscaling* | If your workload is cyclical or unpredictable, enable storage autoscaling to allow RDS to scale up your storage when needed. |
| *Standby instance* | We highly recommend creating a standby instance for production environments.|
| *VPC* | **Select the VPC of your Kubernetes cluster**. Attention: After a database is created, you can't change its VPC. <br> Read more on [VPCs and RDS on Amazon documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_VPC.html).| 
| *Subnet group* | Pick a Subnet group or Create a new one. <br> Read more about [DB Subnet Groups on Amazon documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_VPC.WorkingWithRDSInstanceinaVPC.html#USER_VPC.Subnets). |
| *Public access* | Set the Public access to `No` for production environments. |
| *VPC security group* | Create a new VPC security group and open the postgreSQL port or use an existing one. |
| *Password authentication* or *Password and IAM database authentication* | Choose one or the other. |
| *Database name* | In the *Database options* section, enter Pachyderm's Database name (We are using `pachyderm` in this example.) and click *Create database* to create your PostgreSQL service. Your instance is running. <br>Warning: If you do not specify a database name, Amazon RDS does not create a database.|


!!! Warning "One last step"
      Once your instance is created:

      - You will need to create a second database named "dex" for Pachyderm's authentication service. Note that the database must be named `dex`. Read more about [dex on PostgreSQL on Dex's documentation](https://dexidp.io/docs/storage/#postgres).
      - Additionally, create a new user account and **grant it full CRUD permissions to both `pachyderm` and `dex` databases**. Pachyderm will use the same username to connect to `pachyderm` as well as to `dex`. 

### Update your values.yaml 
Once your databases have been created, add the following fields to your Helm values:


```yaml
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
## 4. Deploy Pachyderm
You have set up your infrastructure, created your S3 bucket, granted your cluster access to your bucket, and, if needed, have configured your EKS cluster to create your pvs.

You can now finalize your values.yaml and deploy Pachyderm.

Note that if you have created an AWS Managed PostgreSQL instance, you will have to replace the Postgresql section below with the appropriate values defined above.
### Update Your Values.yaml   

#### For gp3 EBS Volumes

[Check out our example of values.yaml for gp3](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/examples/aws-gp3-values.yaml) or use our minimal example below.


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

      postgresql:
        persistence:
          storageClass: gp3
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


      postgresql:
        persistence:
          storageClass: gp3
      ```

#### For gp2 EBS Volumes

[Check out our example of values.yaml for gp2](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/examples/aws-gp2-values.yaml) or use our minimal example below.   
    
=== "For Gp2 + Service account annotations"
      ```yaml
      deployTarget: AMAZON
      
      etcd:
        size: 500Gi

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

      postgresql:
        persistence:
          size: 500Gi
      ```  
=== "For Gp2 + AWS Credentials"
      ```yaml
      deployTarget: AMAZON
      
      etcd:
        size: 500Gi

      pachd:
        storage:
          amazon:
            bucket: blah
            region: us-east-2
            # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            id: AKIAIOSFODNN7EXAMPLE
            
            # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html           
            secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

      postgresql:
        persistence:
          size: 500Gi
      ```

Check the [list of all available helm values](../../../../reference/helm_values/) at your disposal in our reference documentation or on [github](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml).

### Deploy Pachyderm On The Kubernetes Cluster


- You can now deploy a Pachyderm cluster by running this command:

      ```shell
      $ helm repo add pach https://helm.pachyderm.com
      $ helm repo update
      $ helm install pachd -f my_values.yaml pach/pachyderm --version <version-of-the-chart>
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
      $ kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m
      ```

      **System Response**

      ```
      pod/pachd-74c5766c4d-ctj82 condition met
      ```

      **Note:** If you see a few restarts on the `pachd` nodes, it means that
      Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
      Kubernetes restarted those pods. You can safely ignore this message.

- Finally, make sure [`pachtl` talks with your cluster](#4-have-pachctl-and-your-cluster-communicate).

## 5. Have 'pachctl' And Your Cluster Communicate

Install [pachctl](../../../../getting_started/local_installation#install-pachctl), then,
assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you are exposing your cluster publicly, retrieve the external IP address of your TCP load balancer or your domain name and:

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      $ echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
      ```

  1. Check that your are using the right context: 

      ```shell
      $ pachctl config get active-context`
      ```

      Your cluster context name should show up.

If you're not exposing `pachd` publicly, you can run:

```shell
# Background this process because it blocks.
$ pachctl port-forward
``` 

## 6. Check That Your Cluster Is Up And Running

```shell
$ pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

