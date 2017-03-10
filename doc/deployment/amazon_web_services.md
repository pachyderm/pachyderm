# Deploying Pachyderm - Amazon Web Services

### Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)

### Deploy Kubernetes

The easiest way to install Kubernetes on AWS is with kops. Kubenetes has provided a [step by step guide](https://github.com/kubernetes/kops/blob/master/docs/aws.md) for the deploy.  Please follow [this guide](https://github.com/kubernetes/kops/blob/master/docs/aws.md) to deploy Kubernetes on AWS.  

Once, you have a Kubernetes cluster up and running in AWS, you should be able to see the following output from `kubectl`:

```shell
$ kubectl get all
NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
svc/kubernetes   10.0.0.1     <none>        443/TCP   22s
```

### Deploy Pachyderm

To deploy Pachyderm we will need to:

1. Add some storage resources on AWS, 
2. Install the Pachyderm CLI tool, `pachctl`, and
3. Deploy Pachyderm on top of the storage resources.

#### Set up the Storage Resources

Pachyderm needs an [S3 bucket](https://aws.amazon.com/documentation/s3/), and a [persistent disk](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html) (EBS) to function correctly.

Here are the environmental variables you should set up to create these resources:

```shell
$ kubectl cluster-info
  Kubernetes master is running at https://1.2.3.4
  ...
$ KUBECTLFLAGS="-s [The public IP of the Kubernetes master. e.g. 1.2.3.4]"

# BUCKET_NAME needs to be globally unique across the entire AWS region
$ BUCKET_NAME=[The name of the S3 bucket where your data will be stored]

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the EBS volume that you are going to create, in GBs. e.g. "10"]

$ AWS_REGION=[the AWS region of your Kubernetes cluster. e.g. "us-west-2" (not us-west-2a)]

$ AWS_AVAILABILITY_ZONE=[the AWS availability zone of your Kubernetes cluster. e.g. "us-west-2a"]

```

Then to actually create the resources, you can run:
```
$ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}

$ aws ec2 create-volume --size ${STORAGE_SIZE} --region ${AWS_REGION} --availability-zone ${AWS_AVAILABILITY_ZONE} --volume-type gp2
```

Record the "volume-id" that is output (e.g. "vol-8050b807") from the above `create-volume` command as shown below (you can also view it in the aws console or with  `aws ec2 describe-volumes`):

```shell
$ STORAGE_NAME=<volume id>
```

Now, as a sanity check, you should be able to see the bucket and the EBS volume that are just created:

```shell
aws s3api list-buckets --query 'Buckets[].Name'
aws ec2 describe-volumes --query 'Volumes[].VolumeId'
```

#### Install `pachctl`

To deploy and interact with Pachyderm, you will need `pachctl`, a command-line utility used for Pachyderm. To install `pachctl` run one of the following:


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.2
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

#### Deploy Pachyderm

When you installed kops, you should have created a dedicated IAM user (see [here](https://github.com/kubernetes/kops/blob/master/docs/aws.md#aws) for details).  To deploy Pachyderm you will need to export these credentials to the following environmental variables:

```sh
$ AWS_ID=[access key ID]

$ AWS_KEY=[secret access key]
```
Run the following command to deploy your Pachyderm cluster:

```shell
$ pachctl deploy amazon ${BUCKET_NAME} ${AWS_ID} ${AWS_KEY} ${AWS_TOKEN} ${AWS_REGION} ${STORAGE_NAME} ${STORAGE_SIZE}
```

Note: For a permanent deployment, all you have to do is leave the token blank and make sure the user has the right permissions.

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME               READY     STATUS    RESTARTS   AGE
po/etcd-xzc0d      1/1       Running   0          55s
po/pachd-6m6wm     1/1       Running   0          55s
po/rethink-388b3   1/1       Running   0          55s

NAME         DESIRED   CURRENT   READY     AGE
rc/etcd      1         1         1         55s
rc/pachd     1         1         1         55s
rc/rethink   1         1         1         55s

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                                          AGE
svc/etcd         10.0.0.92    <none>        2379/TCP,2380/TCP                                55s
svc/kubernetes   10.0.0.1     <none>        443/TCP                                          9m
svc/pachd        10.0.0.61    <nodes>       650:30650/TCP,651:30651/TCP                      55s
svc/rethink      10.0.0.87    <nodes>       8080:32080/TCP,28015:32081/TCP,29015:32085/TCP   55s

NAME              DESIRED   SUCCESSFUL   AGE
jobs/pachd-init   1         1            55s

```
Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.
```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.2
pachd               1.3.2
```

## Microsoft Azure

### Prerequisites

* Install [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) >= 0.10.6
* Install [jq](https://stedolan.github.io/jq/download/)

### Deploy Kubernetes

The easiest way to deploy a Kubernetes cluster is to use the [official Kubernetes guide](http://kubernetes.io/docs/getting-started-guides/azure/).

### Deploy Pachyderm

#### Set up the Storage Infrastructure

Pachyderm requires an object store ([Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)) and a [data disk](https://azure.microsoft.com/documentation/articles/virtual-machines-windows-about-disks-vhds/#data-disk) to function correctly.

Here are the parameters required to create these resources:

```sh
# Needs to be globally unique across the entire Azure location
$ AZURE_RESOURCE_GROUP=[The name of the resource group where the Azure resources will be organized]

$ AZURE_LOCATION=[The Azure region of your Kubernetes cluster. e.g. "West US2"]

# Needs to be globally unique across the entire Azure location
$ AZURE_STORAGE_NAME=[The name of the storage account where your data will be stored]

$ CONTAINER_NAME=[The name of the Azure blob container where your data will be stored]

# Needs to end in a ".vhd" extension
$ STORAGE_NAME=pach-disk.vhd

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the data disk volume that you are going to create, in GBs. e.g. "10"]
```

And then run:

```sh
$ azure group create --name ${AZURE_RESOURCE_GROUP} --location ${AZURE_LOCATION}
$ azure storage account create ${AZURE_STORAGE_NAME} --location ${AZURE_LOCATION} --resource-group ${AZURE_RESOURCE_GROUP} --sku-name LRS --kind Storage

# Retrieve the Azure Storage Account Key
$ AZURE_STORAGE_KEY=`azure storage account keys list ${AZURE_STORAGE_NAME} --resource-group ${AZURE_RESOURCE_GROUP} --json | jq .[0].value -r`

# Build the microsoft_vhd container.
$ make docker-build-microsoft-vhd

# Create an empty data disk in the "disks" container
$ STORAGE_VOLUME_URI=`docker run -it microsoft_vhd ${AZURE_STORAGE_NAME} ${AZURE_STORAGE_KEY} "disks" ${STORAGE_NAME} ${STORAGE_SIZE}G`
```

To check that everything has been setup correctly, try:

```sh
$ azure storage account list
# should see a number of storage accounts, including the one specified with ${AZURE_STORAGE_NAME}

$ azure storage blob list --account-name ${AZURE_STORAGE_NAME} --account-key ${_AZURE_STORAGE_KEY}
# should see a disk with the name ${STORAGE_NAME}
```

#### Install Pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

```sh
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.2
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

#### Start Pachyderm

Now we're ready to boot up Pachyderm:

```sh
$ pachctl deploy microsoft ${CONTAINER_NAME} ${AZURE_STORAGE_NAME} ${AZURE_STORAGE_KEY} ${STORAGE_VOLUME_URI} ${STORAGE_SIZE}
```

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME               READY     STATUS    RESTARTS   AGE
po/etcd-xzc0d      1/1       Running   0          55s
po/pachd-6m6wm     1/1       Running   0          55s
po/rethink-388b3   1/1       Running   0          55s

NAME         DESIRED   CURRENT   READY     AGE
rc/etcd      1         1         1         55s
rc/pachd     1         1         1         55s
rc/rethink   1         1         1         55s

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                                          AGE
svc/etcd         10.0.0.92    <none>        2379/TCP,2380/TCP                                55s
svc/kubernetes   10.0.0.1     <none>        443/TCP                                          9m
svc/pachd        10.0.0.61    <nodes>       650:30650/TCP,651:30651/TCP                      55s
svc/rethink      10.0.0.87    <nodes>       8080:32080/TCP,28015:32081/TCP,29015:32085/TCP   55s

NAME              DESIRED   SUCCESSFUL   AGE
jobs/pachd-init   1         1            55s

```
Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl portforward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.2
pachd               1.3.2
```


## OpenShift

[OpenShift](https://www.openshift.com/) is a popular enterprise Kubernetes distribution.  Pachyderm can run on OpenShift with two additional steps:

1. Make sure that privilege containers are allowed (they are not allowed by default):  `oc edit scc` and set `allowPrivilegedContainer: true` everywhere.
2. Remove `hostPath` everywhere from your cluster manifest (e.g. `etc/kube/pachyderm-versioned.json` if you are deploying locally).

Problems related to OpenShift deployment are tracked in this issue: https://github.com/pachyderm/pachyderm/issues/336

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.
