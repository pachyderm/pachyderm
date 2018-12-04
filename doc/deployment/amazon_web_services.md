# Amazon Web Services

## Advanced

- [Deploy within an existing VPC](https://github.com/pachyderm/pachyderm/blob/master/doc/deployment/amazon_web_services/existing_vpc.md)
- [Connect to your Pachyderm Cluster](https://github.com/pachyderm/pachyderm/blob/master/doc/deployment/amazon_web_services/connecting_to_your_cluster.md)

## Standard Deployment

We recommend one of the following two methods for deploying Pachyderm on AWS:

2. [By manually deploying Kubernetes and Pachyderm.](amazon_web_services.html#manual-pachyderm-deploy)
    - This is appropriate if you (i) already have a kubernetes deployment, (ii) if you would like to customize the types of instances, size of volumes, etc. in your cluster, (iii) if you're setting up a production cluster, or (iv) if you are processing a lot of data or have computationally expensive workloads.
1. [By executing a one shot deploy script that will both deploy Kubernetes and Pachyderm.](amazon_web_services.html#one-shot-script)
    - This option is appropriate if you are just experimenting with Pachyderm. The one-shot script will get you up and running in no time!

In addition, we recommend setting up AWS CloudFront for any production deployments. AWS puts S3 rate limits in place that can limit the data throughput for your cluster, and CloudFront helps mitigate this issue. Follow these instructions to deploy with CloudFront

- [Deploy a Pachyderm cluster with CloudFront](./aws_cloudfront.html)

## Manual Pachyderm Deploy

### Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)
- [pachctl](#install-pachctl)

### Deploy Kubernetes

The easiest way to install Kubernetes on AWS (currently) is with kops. You can follow [this step-by-step guide from Kubernetes](https://github.com/kubernetes/kops/blob/master/docs/aws.md) for the deploy. Note, we recommend using at `r4.xlarge` or larger instances in the cluster.  

Once, you have a Kubernetes cluster up and running in AWS, you should be able to see the following output from `kubectl`:

```shell
$ kubectl get all
NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
svc/kubernetes   ClusterIP   100.64.0.1   <none>        443/TCP   7m
```

### Deploy Pachyderm

To deploy Pachyderm on your k8s cluster you will need to:

1. Install the `pachctl` CLI tool,
2. Add some storage resources on AWS,
3. Deploy Pachyderm on top of the storage resources.

#### Install `pachctl`

To deploy and interact with Pachyderm, you will need `pachctl`, Pachyderm's command-line utility. To install `pachctl` run one of the following:


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.8

# For Linux (64 bit) or Window 10+ on WSL:
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.8.0/pachctl_1.8.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version --client-only` to verify that `pachctl` has been successfully installed.

```sh
$ pachctl version --client-only
1.7.0
```

#### Set up the Storage Resources

Pachyderm needs an [S3 bucket](https://aws.amazon.com/documentation/s3/), and a [persistent disk](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html) (EBS in AWS) to function correctly.

Here are the environmental variables you should set up to create and utilize these resources:

```shell
# BUCKET_NAME needs to be globally unique across the entire AWS region
$ BUCKET_NAME=<The name of the S3 bucket where your data will be stored>

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=<the size of the EBS volume that you are going to create, in GBs. e.g. "10">

$ AWS_REGION=<the AWS region of your Kubernetes cluster. e.g. "us-west-2" (not us-west-2a)>
```

Then to actually create the backing S3 bucket, you can run one of the following:
```
# If AWS_REGION is us-east-1.
$ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}

# If AWS_REGION is outside of us-east-1.
$ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
```

As a sanity check, you should be able to see the bucket that you just created when you run the following:

```shell
$ aws s3api list-buckets --query 'Buckets[].Name'
```

#### Deploy Pachyderm

You can deploy Pachyderm on AWS using:

- [An IAM role](#deploying-with-an-iam-role), or
- [Static credentials](#deploying-with-static-credentials)

##### Deploying with an IAM role

Run the following command to deploy your Pachyderm cluster:

```shell
$ pachctl deploy amazon ${BUCKET_NAME} ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --iam-role <your-iam-role>
```

Note that for this to work, the following need to be true:

* The nodes on which Pachyderm is deployed need to be assigned with the utilized IAM role.  If you created your cluster with `kops`, the nodes should have a dedicated IAM role.  You can find this IAM role by going to the AWS console, clicking on one of the EC2 instance in the k8s cluster, and inspecting the "Description" of the instance.

* The IAM role needs to have access to the bucket you just created. To ensure that it has access, you can go to the `Permissions` tab of the IAM role and edit the policy to include the following segment (Make sure to replace `your-bucket` with your actual bucket name):

    ```json
    {
        "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::your-bucket"
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
            "arn:aws:s3:::your-bucket/*"
        ]
    }
    ```

* The IAM role needs to have the proper "trust relationships" set up.  You can verify this by navigating to the `Trust relationships` tab of your IAM role, clicking `Edit trust relationship`, and ensuring that you see a `statement` with `sts:AssumeRole`.  For instance, this would be a valid trust relationship:

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

Once you've run `pachctl deploy ...` and waited a few minutes, you should see the following running pods in Kubernetes:

```sh
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
etcd-0                   1/1       Running   0          4m
pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
```

**Note**: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before etcd was ready so it restarted them.

If you see the above pods running, the last thing you need to do is forward a couple ports so that `pachctl` can talk to the cluster:

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can verify that the cluster is working by executing `pachctl version`, which should return a version for both `pachctl` and `pachd`:

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.0
pachd               1.7.0
```

##### Deploying with static credentials

When you installed kops, you should have created a dedicated IAM user (see [here](https://github.com/kubernetes/kops/blob/master/docs/aws.md#aws) for details).  You could deploy Pachyderm using the credentials of this IAM user directly, although that's not recommended:

```sh
$ AWS_ACCESS_KEY_ID=<access key ID>

$ AWS_SECRET_ACCESS_KEY=<secret access key>
```

Run the following command to deploy your Pachyderm cluster:

```shell
$ pachctl deploy amazon ${BUCKET_NAME} ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --credentials "${AWS_ACCESS_KEY_ID},${AWS_SECRET_ACCESS_KEY},"
```

Note, the `,` at the end of the `credentials` flag in the deploy command is for an optional temporary AWS token. You might utilize this sort of temporary token if you are just experimenting with a deploy. However, such a token should NOT be used for a production deploy.

It may take a few minutes for Pachyderm to start running on the cluster, but you you should eventually see the following running pods:

```sh
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
etcd-0                   1/1       Running   0          4m
pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
```

If you see an output similar to the above, the last thing you need to do is forward a couple ports so that `pachctl` can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can verify that the cluster is working by running `pachctl version`, which should return a version for both `pachctl` and `pachd`:
```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.0
pachd               1.7.0
```

## One Shot Script

### Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)
- [pachctl](#install-pachctl)
- [jq](https://stedolan.github.io/jq/download/)
- [uuid](http://man7.org/linux/man-pages/man1/uuidgen.1.html)

### Run the deploy script

Once you have the prerequisites mentioned above, download and run our AWS deploy script by running:

```
$ curl -o aws.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/aws.sh
$ chmod +x aws.sh
$ sudo -E ./aws.sh
```

This script will use `kops` to deploy Kubernetes and Pachyderm in AWS.  The script will ask you for your AWS credentials, region preference, etc.  If you would like to customize the number of nodes in the cluster, node types, etc., you can open up the deploy script and modify the respective fields.

The script will take a few minutes, and Pachyderm will take an addition couple of minutes to spin up.  Once it is up, `kubectl get pods` should return something like:

```sh
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
etcd-0                   1/1       Running   0          4m
pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
```

### Connect `pachctl`

You will then need to forward a couple ports so that `pachctl` can talk to the cluster:

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can verify that the cluster is working by executing `pachctl version`, which should return a version for both `pachctl` and `pachd`:

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.0
pachd               1.7.0
```

### Remove

You can delete your Pachyderm cluster using `kops`:

```sh
$ kops delete cluster
```

In addition, there is the entry in `/etc/hosts` pointing to the cluster that will need to be manually removed. Similarly, kubernetes state s3 bucket and pachyderm storage bucket will need to be manually removed.
