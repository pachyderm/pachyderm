# Deploy Pachyderm on AWS

Once your Kubernetes cluster is up,
you are ready to deploy Pachyderm.

Complete the following steps:

1. [Create an S3 bucket](#create-an-S3-object-store-bucket-for-data) for your data and grant Pachyderm access.
1. [Enable Persistent Volumes Creation](#2-enable-your-persistent-volumes-creation)
1. [Deploy Pachyderm ](#3-deploy-pachyderm)
1. Finally, you will need to install [pachctl](../../../../getting_started/local_installation#install-pachctl) to [interact with your cluster]((#have-pachctl-and-your-cluster-communicate)).
1. And check that your cluster is [up and running](#5-check-that-your-cluster-is-up-and-running)

## 1- Create an S3 bucket
### Create an **S3 object store bucket for data**
!!! Warning
      The S3 bucket name must be globally unique across the whole
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

## 2- Enable Your Persistent Volumes Creation

etcd and PostgreSQL (metadata storage) each claim the creation of a pv. 

!!! Important
      The metadata services generally require a small persistent volume size (i.e. 10GB) **but high IOPS (1500)**.
      Note that Pachyderm out-of-the-box deployment comes with **gp2** default EBS volumes. 
      While it might be easier to set up for test or development environments, **we highly recommend to use SSD gp3 in production**. A **gp3** EBS volume delivers a baseline performance of 3,000 IOPS and 125MB/s at any volume size. Any other disk choice may require to **oversize the volume significantly to ensure enough IOPS**.

      See [volume types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html).

If you plan on using **gp2** EBS volumes, [skip this section and jump to the deployment of Pachyderm](#3-deploy-pachyderm).
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
## 3- Deploy Pachyderm
You have created your S3 bucket, given your cluster access to your bucket, and have configured your EKS cluster to create your pvs (metadata).

You can now deploy Pachyderm.
### Create Your Values.yaml   

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
              eks.amazonaws.com/role-arn: arn:aws:iam::190146978412:role/pachyderm-bucket-access

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

            # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            id: AKIAIOSFODNN7EXAMPLE

            # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
            region: us-east-2

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
            
            # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
            id: AKIAIOSFODNN7EXAMPLE
            
            # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html           
            secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
            region: us-east-2

      postgresql:
        persistence:
          size: 500Gi
      ```

Check the [list of all available helm values](../../../../reference/helm_values/) at your disposal in our reference documentation.

!!! Important "Load Balancer Setup" 
      If you would like to expose your pachd instance to the internet via load balancer, add the following config under `pachd` to your `values.yaml`
      ```yaml
       pachd:
        service:
         type: LoadBalancer
      ```

      **NOTE: It is strongly recommended to configure SSL when exposing Pachyderm publicly.**

### Deploy Pachyderm On The Kubernetes Cluster

Refer to our generic ["Helm Install"](./helm_install.md) page for more information on the required installations and modus operandi of an installation using `Helm`.

- Now you can deploy a Pachyderm cluster by running this command:

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

## 4- Have 'pachctl' And Your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you specified `LoadBalancer` in the `values.yaml` file:

  1. Retrieve the external IP address of the service.  When listing your services again, you should see an external IP address allocated to the `pachd` service 

      ```shell
      $ kubectl get service
      ```

  1. Update the context of your cluster with their direct url, using the external IP address above:

      ```shell
      $ echo '{"pachd_address": "grpc://<external-IP-address>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
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

## 5- Check That Your Cluster Is Up And Running

```shell
$ pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```



## RDS - AWS managed PostgreSQL database

Tutorial https://aws.amazon.com/getting-started/hands-on/create-connect-postgresql-db/

 - In ASW Management Console, find RDS under database and open the RDS console
 - On the top right, select the region matching Your Pachyderm cluster
 - In the Create database section, choose Create database.
 - Choose the PostgreSQL icon and select the latest version  (Which??? How do I find the compatibility Pachyderm version - postgreSQL version)
 - You will now configure your DB instance:
 Settings:
  <DB instance identifier>, <Master username>, <Master password>
 Instance specifications:
    - <DB instance class>: Default standard should work  - You can change the instance type later on to optimize your performances/costs
    - Storage:
    < Storage Type> if gp2 point warning for iops and oversize the disk accordingly (allocated storage  greater than 1TB)
                  if io1 keep the 100 GiB default

    <Enable storage autoscaling> optionnal If your workload is cyclical or unpredictable, you would enable storage autoscaling to enable RDS to automatically scale up your storage when needed. 

    -Availability & durability:
    We hightly recommend creating a stanby instance for production environments

    -Connectivity
    !!! Warning
      Select the VPC of your Kubernetes cluster.
      After a database is created, you can't change its VPC.
      <Subnet group> Pick a subnet or Create a new one
      <Public access> No Then how do you connect with pgAdmin?
      <VPC security group> Check back with Dale - Create a new security group and open the postgres port???? or use existing

     - Database authentication
      Either Password authentication or Password and IAM database authentication

      - Additional configuration
      In Database options, enter a database name. If you do not specify a database name, Amazon RDS does not create a database.

      Click **Create database** to create your postgres service.

      Additionnally, you need to create a second schema named "dex" for the authentication service and grant (which?) privileges to whom? https://dexidp.io/docs/storage/#postgres

Add the following fields to your Helm values

```yaml
global:
  postgresql:
    postgresqlUsername: "pachyderm"
    postgresqlPassword: "pachyderm" 
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
'''