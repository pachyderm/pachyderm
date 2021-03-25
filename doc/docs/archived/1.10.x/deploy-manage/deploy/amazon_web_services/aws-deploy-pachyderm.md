# Deploy Pachyderm on AWS

After you deploy Kubernetes cluster by using `kops` or `eksctl`,
you can deploy Pachyderm on top of that cluster.

You need to complete the following steps to deploy Pachyderm:

1. Install `pachctl` as described in [Install pachctl](../../../../getting_started/local_installation#install-pachctl).
1. Add stateful storage for Pachyderm as described in [Add Stateful Storage](#add-stateful-storage).
1. Deploy Pachyderm by using an [IAM role](#deploy-pachyderm-with-an-iam-role)
(recommended) or [an access key](#deploy-pachyderm-with-an-access-key).

## Add Stateful Storage

Pachyderm requires the following types of persistent storage:

1. An S3 object store bucket for data. The S3 bucket name
  must be globally unique across the whole
  Amazon region. Therefore, add a descriptive prefix to the S3 bucket
  name, such as your username.

1. An Elastic Block Storage (EBS) persistent volume (PV) for Pachyderm metadata.
  Pachyderm recommends that you assign at least 10 GB of space and at least 1500
  IOPS (which requires provisioning at least 500 GB of space on gp2 volumes, the
  default [volume type](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html))
  for this persistent EBS volume. If you expect your cluster to be very long
  running or scale to thousands of jobs per commits, you might need to go add
  more storage.  However, you can easily increase the size of the persistent
  volume later.

To add stateful storage, complete the following steps:

1. Set up the following system variables:

   * `BUCKET_NAME` — A globally unique S3 bucket name.
   * `STORAGE_SIZE` — The size of the persistent volume in GB. For example, `500`.
   * `AWS_REGION` — The AWS region of your Kubernetes cluster. For example,
   `us-west-2` and not `us-west-2a`.
  

1. Create an S3 bucket:

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

1. Verify that the S3 bucket was created:

   ```
   aws s3api list-buckets --query 'Buckets[].Name'
   ```

### (Optional) Set up Bucket Encryption

Amazon S3 supports two types of bucket encryption — server-side encryption
(SSE-S3) and AWS Key Management Service (AWS KMS), which stores customer
master keys. Pachyderm supports both these methods. Therefore, when you
are creating a bucket for your Pachyderm cluster, you can set up either
of them. Because Pachyderm requests to buckets do not include encryption
information, the method that you select for the bucket is applied.
Setting up communication between Pachyderm object storage clients and AWS KMS
to append encryption information to Pachyderm requests is not supported and
not recommended.

To set up bucket encryption, see [Amazon S3 Default Encryption for S3 Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html).

## Deploy Pachyderm with an IAM Role

IAM roles provide better user management and security
capabilities compared to access keys. If a malicious user gains access to
an access key, your data might become compromised. Therefore, enterprises
often opt out to use IAM roles rather than access keys for production
deployments.

You need to configure the following IAM settings:

* The worker nodes on which Pachyderm is deployed must be associated
with the IAM role that is assigned to the Kubernetes cluster.
If you created your cluster by using `kops` or `eksctl`
the nodes must have a dedicated IAM role already assigned.

* The IAM role must have access to the S3 bucket that you created for
Pachyderm.

* The IAM role must have correct trust relationships.

  You need to set a system variable `IAM_ROLE` to the name
  of the IAM role that you will use to deploy the cluster.
  This role is different from the Role ARN or the Instance
  Profile ARN of the role. It is the actual role name.

To deploy Pachyderm with an IAM role, complete the following steps:

1. Find the IAM role assigned to the cluster:

   1. Go to the AWS Management console.
   1. Select an EC2 instance in the Kubernetes cluster.
   1. Click **Description**.
   1. Find the **IAM Role** field.

1. Enable access to the S3 bucket for the IAM role:

   1. In the **IAM Role** field, click on the IAM role.
   1. In the **Permissions** tab, click **Edit policy**.
   1. Select the **JSON** tab.
   1. Append the following text to the end of the existing JSON:

      ```json
      {
          "Effect": "Allow",
              "Action": [
                  "s3:ListBucket"
              ],
              "Resource": [
                  "arn:aws:s3:::<your-bucket>"
              ]
      },
      {
          "Effect": "Allow",
          "Action": [
              "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
          ],
          "Resource": [
              "arn:aws:s3:::<your-bucket>/*"
          ]
      }
      ```

      Replace `<your-bucket>` with the name of your S3 bucket.

      **Note:** For the EKS cluster, you might need to use the
      **Add inline policy** button and create a name for the new policy.
      The JSON above is inserted between the square brackets for the `Statement` element.

1. Set up trust relationships for the IAM role:

   1. Click the **Trust relationships > Edit trust relationship**.
   1. Ensure that you see a statement with `sts:AssumeRole`. Example:

      ```json
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Principal": {
              "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
          }
        ]
      }
      ```

1. Set the system variable `IAM_ROLE` to the IAM role name
   for the Pachyderm deployment.

1. Deploy Pachyderm:

   ```shell
   pachctl deploy amazon ${BUCKET_NAME} ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --iam-role ${IAM_ROLE}
   ```

   The deployment takes some time. You can run `kubectl get pods` periodically
   to check the status of deployment. When Pachyderm is deployed, the command
   shows all pods as `READY`:

   ```shell
   kubectl get pods
   ```

   **System Response:**

   ```shell
   NAME                     READY     STATUS    RESTARTS   AGE
   dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
   etcd-0                   1/1       Running   0          4m
   pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
   ```

   **Note:** If you see a few restarts on the `pachd` nodes, it means that
   Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
   Kubernetes restarted those pods. You can safely ignore this message.

1. Verify that the Pachyderm cluster is up and running:

   ```shell
   pachctl version
   ```

   **System Response:**

   ```shell
   COMPONENT           VERSION
   pachctl             1.9.7
   pachd               1.9.7
   ```

   * If you want to access the Pachyderm UI or use the S3 gateway, you need to
   forward Pachyderm ports. Open a new terminal window and run the
   following command:

   ```shell
   pachctl port-forward
   ```

## Deploy Pachyderm with an Access Key

When you installed `kops`, you created a dedicated IAM
user with access credentials such as an access key and
secret key. You can deploy
Pachyderm by using the credentials of this IAM user
directly. However, deploying Pachyderm with an
access key might not satisfy your enterprise security
requirements. Therefore, deploying with an IAM role
is preferred.

To deploy Pachyderm with an access key, complete the following
steps:

1. Run the following command to deploy your Pachyderm cluster:

   ```shell
   pachctl deploy amazon ${BUCKET_NAME} ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --credentials "${AWS_ACCESS_KEY_ID},${AWS_SECRET_ACCESS_KEY},"
   ```

   The `,` at the end of the `credentials` flag in the deploy
   command is for an optional temporary AWS token. You might use
   such a token if you are just experimenting with
   Pachyderm. However, do not use this token in a
   production deployment.

   The deployment takes some time. You can run `kubectl get pods` periodically
   to check the status of deployment. When Pachyderm is deployed, the command
   shows all pods as `READY`:

    ```shell
    kubectl get pods
    ```

    **System Response:**

    ```shell
    NAME                     READY     STATUS    RESTARTS   AGE
    dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
    etcd-0                   1/1       Running   0          4m
    pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
    ```

    **Note:** If you see a few restarts on the `pachd` nodes, it means that
    Kubernetes tried to bring up those pods before `etcd` was ready.
    Therefore, Kubernetes restarted those pods. You can safely ignore this
    message.

1. Verify that the Pachyderm cluster is up and running:

   ```shell
   pachctl version
   ```

   **System Response:**

   ```shell

   COMPONENT           VERSION
   pachctl             1.9.7
   pachd               1.9.7
   ```

   * If you want to access the Pachyderm UI or use S3 gateway, you need to
   forward Pachyderm ports. Open a new terminal window and run the
   following command:

     ```shell
     pachctl port-forward
     ```
