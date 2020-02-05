# Deploy Into an Existing VPC

## Prereqs

- Terraform
- An existing AWS VPC deployed

## How to generate terraform k8s cluster deployment manifest

This how to is based off of [this guide](https://ryaneschinger.com/blog/kubernetes-aws-vpc-kops-terraform/)

1) Collect the following info / set the following env vars

```
VPC_ID=vpc-2345
ZONE_ID=34l5kj34l5kj
ZONES=us-east-1a,us-east-1b,us-east-1c
# the cluster name will also be its domain
# and needs to be a valid subdomain on the hosted zone
NAME=prod.sourceai.io 
```

Collect your list of subnets, which should look like this:

```
  subnets:
  - egress: nat-sdfgsdfgsdfgsdfg
    id: subnet-2345bc2345b
    name: us-east-1a
    type: Private
    zone: us-east-1a
  - egress: nat-sdfgsdfgsdfgsdfg
    id: subnet-57b3575b375b
    name: us-east-1b
    type: Private
    zone: us-east-1b
  - egress: nat-sdfgsdfgsdfgsdfg
    id: subnet-0879ef078ef087
    name: us-east-1c
    type: Private
    zone: us-east-1c
  - id: subnet-2263be6e26be62
    name: Public Subnet 1
    type: Utility
    zone: us-east-1a
  - id: subnet-3444b5425b5
    name: Public Subnet 2
    type: Utility
    zone: us-east-1b
  - id: subnet-2314c334c43
    name: Public Subnet 3
    type: Utility
    zone: us-east-1c
```

Note: For some silly reason, the private subnet names need to match the zone.
This seems to be a requirement for the change to be accepted by kops.

2) Create a kops state store bucket

We need to do this because we're going to use kops to stage the changes, then
emit them as a terraform manifest. To do that kops needs a state store.

You can use TF to generate an s3 bucket. [For example](https://github.com/ryane/kubernetes-aws-vpc-kops-terraform/blob/master/main.tf#L44). Otherwise, here's one
way to do it w some basic error handling:

```
create_s3_bucket() {
  if [[ "$#" -lt 1 ]]; then
    echo "Error: create_s3_bucket needs a bucket name"
    return 1
  fi
  BUCKET="${1#s3://}"

  # For some weird reason, s3 emits an error if you pass a location constraint when location is "us-east-1"
  if [[ "${AWS_REGION}" == "us-east-1" ]]; then
    aws s3api create-bucket --bucket ${BUCKET} --region ${AWS_REGION}
  else
    aws s3api create-bucket --bucket ${BUCKET} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
  fi
}

export AWS_REGION="us-east-1"

create_s3_bucket some_bucket_name
```

3) Create the kops cluster

```
kops create cluster \
     --master-zones $ZONES \
     --zones $ZONES \
     --topology private \
     --dns-zone $ZONE_ID \
     --networking calico \
     --vpc $VPC_ID \
     --target=terraform \
     --out=. \
     ${NAME}
```

4) Update the kops cluster

First edit the deployment to specify your VPC, CIDR, and subnets:

```
kops edit cluster $NAME
```

You can find the CIDR listed on the AWS console.


Then update the cluster:

```
kops update cluster \
   --out=. \
   --target=terraform \
   ${NAME}
```

Which applies the changes to the kops state store and stages them there.

5) Deploy using terraform

```
terraform plan
terraform apply
```

6) Tear down

To tear down, do:

```
terraform destroy
kops delete cluster $NAME
```

## How to generate k8s Pachyderm cluster manifest

The [Deploy Pachyderm on Amazon AWS](https://docs.pachyderm.com/latest/deploy-manage/deploy/amazon_web_services/)
section provides an overview of Pachyderm cluster manifest generation.

But it boils down to this.

1) Create an s3 bucket for the data store
2) Set the `BUCKET_NAME`, `STORAGE_SIZE`, `AWS_REGION`, and AWS credentials env
vars
3) Run the `pachctl deploy amazon ...` command w the `--dry-run` flag to emit
the yaml k8s manifest
4) Store that manifest in the infra repo
5) Deploy via `kubectl create -f pachyderm.yaml`


## Next Steps


[Connect to your Pachyderm Cluster](connecting_to_your_cluster.html)
