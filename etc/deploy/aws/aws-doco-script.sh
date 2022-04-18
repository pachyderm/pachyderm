#!/bin/bash

# Preq: Run aws config to configure auth context profile before running this script.

set -xeou pipefail

jq --version || echo "This script requires jq"
psql --version || echo "This script requires psql"

CLUSTER_NAME="acerbic-aardvark"
POSTGRES_SQL_DB_USER_PASSWORD="correcthorsebatterystaple"
AWS_REGION="us-east-2"
AWS_PROFILE="default"


EMPTY_CLUSTER=false
print-help ()
{
    echo "$0: Create an AWS cluster" >&2
    echo
    echo "Options:"
    echo
    echo "-h  display this help text"
    echo "-e  create the cluster without installing Pachyderm"
    echo
    exit
}
while getopts he- OPT; do
    case "$OPT" in
	h )   set +x; print-help ;;
	e )   EMPTY_CLUSTER=true ;;
	??* ) close "Invalid option --$OPT" ;;
	? )   exit 2 ;;
    esac
done

GLOBAL_TAG="${CLUSTER_NAME}"
BUCKET_NAME="${CLUSTER_NAME}-bucket"
S3_BUCKET_IAM_POLICY_NAME="${CLUSTER_NAME}-bucket-iam-policy"
S3_BUCKET_IAM_ROLE_NAME="${CLUSTER_NAME}-bucket-iam-role"
PODS_BUCKET_ACCESS_IAM_SA="${CLUSTER_NAME}-bucket-access-iam-sa"
EBS_CSI_DRIVER_POLICY_NAME="${CLUSTER_NAME}-EBS-CSI-Driver-Policy"
POSTGRES_SQL_ID="${CLUSTER_NAME}-postgresql"
DB_SUBNET_GROUP_NAME="${CLUSTER_NAME}-db-public-subnet-group"
POSTGRES_SQL_DB_NAME_1="pachyderm"
POSTGRES_SQL_DB_USER_NAME="master"
POSTGRES_SQL_DB_APP_USER_NAME="pachyderm"


# Create cluster
eksctl create cluster --name "${CLUSTER_NAME}" --region "${AWS_REGION}" --profile "${AWS_PROFILE}" --tags "Key=Name,Value=${GLOBAL_TAG}"

# Verify deployment
kubectl get all

# Create an S3 object store bucket for data
aws s3api create-bucket --bucket "${BUCKET_NAME}" --region "${AWS_REGION}" --create-bucket-configuration "LocationConstraint=${AWS_REGION}"

# Verify that the S3 bucket was created
aws s3 ls

# Create an IAM OIDC identity provider for cluster
# View cluster's OIDC provider URL.
OIDC_PROVIDER_URL=$(aws eks describe-cluster --name ${CLUSTER_NAME} --query "cluster.identity.oidc.issuer" --output text --region "${AWS_REGION}")

eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} --approve --region "${AWS_REGION}"

# Create an IAM policy that gives access to bucket:
cat <<EOF > policy.json
{
   "Version":"2012-10-17",
   "Statement":[
      {
         "Effect":"Allow",
         "Action":[
            "s3:ListBucket"
         ],
         "Resource":[
            "arn:aws:s3:::$BUCKET_NAME"
         ]
      },
      {
         "Effect":"Allow",
         "Action":[
            "s3:PutObject",
            "s3:GetObject",
            "s3:DeleteObject"
         ],
         "Resource":[
            "arn:aws:s3:::$BUCKET_NAME/*"
         ]
      }
   ]
}
EOF

# Create managed policy and capture policy's Amazon resource name (ARN)
MANAGED_POLICY_ARN=$(aws iam create-policy \
    --policy-name "${S3_BUCKET_IAM_POLICY_NAME}" \
    --policy-document file://policy.json \
    --tags "Key=Name,Value=${GLOBAL_TAG}" \
    --output json | jq -r '.Policy.Arn')


#Create an IAM service account with the policy attached.
ACCOUNTID=$(aws sts get-caller-identity --query Account --output text)
eksctl create iamserviceaccount \
    --name "${PODS_BUCKET_ACCESS_IAM_SA}" \
    --cluster "${CLUSTER_NAME}" \
    --attach-policy-arn "arn:aws:iam::${ACCOUNTID}:policy/${S3_BUCKET_IAM_POLICY_NAME}" \
    --approve \
    --override-existing-serviceaccounts \
    --region "${AWS_REGION}" \
    --tags "Key=Name,Value=${GLOBAL_TAG}"
    

# Create an IAM role as a Web Identity using the cluster OIDC procider as the identity provider.
OPENID_CONNECT_PROVIDER_ARN=$(aws iam list-open-id-connect-providers --output json | grep "${OIDC_PROVIDER_URL##*/}" | awk '{print $2}')

cat <<EOF > Test-Role-Trust-Policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Principal": {
                "Federated": $OPENID_CONNECT_PROVIDER_ARN
            },
            "Condition": {
                "StringEquals": {
                    "${OIDC_PROVIDER_URL#https://}:aud": [
                        "sts.amazonaws.com"
                    ]
                }
            }
        }
    ]
}
EOF

ROLE_ARN=$(aws iam create-role --role-name "${S3_BUCKET_IAM_ROLE_NAME}" --assume-role-policy-document file://Test-Role-Trust-Policy.json --tags "Key=Name,Value=${GLOBAL_TAG}" --output json | jq -r '.Role.Arn')

# Attach a managed policy to an IAM role
aws iam attach-role-policy --policy-arn "${MANAGED_POLICY_ARN}" --role-name "${S3_BUCKET_IAM_ROLE_NAME}"

# (Optional) Set Up Bucket Encryption
# TODO: Set up bucket encryption

ACCOUNTID=$(aws sts get-caller-identity --query Account --output text)

curl -o example-iam-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-ebs-csi-driver/master/docs/example-iam-policy.json

aws iam create-policy \
    --policy-name ${EBS_CSI_DRIVER_POLICY_NAME} \
    --policy-document file://example-iam-policy.json \
    --tags Key=Name,Value=${GLOBAL_TAG}
    

eksctl create iamserviceaccount \
    --name ebs-csi-controller-sa \
    --namespace kube-system \
    --cluster ${CLUSTER_NAME} \
    --attach-policy-arn "arn:aws:iam::${ACCOUNTID}:policy/${EBS_CSI_DRIVER_POLICY_NAME}" \
    --approve \
    --role-only \
    --region "${AWS_REGION}" \
    --tags Key=Name,Value=${GLOBAL_TAG}

# aws cloudformation get created stack name
STACK_NAME=$(aws cloudformation describe-stacks --output json --region "${AWS_REGION}" | jq -r '.Stacks[].StackName' | grep "${CLUSTER_NAME}" | grep 'ebs-csi')

CREATED_ROLE_NAME=$(aws cloudformation describe-stack-resources --query 'StackResources[0].PhysicalResourceId' --output text --stack-name "${STACK_NAME}" --region "${AWS_REGION}")

eksctl create addon --name aws-ebs-csi-driver --cluster "${CLUSTER_NAME}" --service-account-role-arn "arn:aws:iam::${ACCOUNTID}:role/${CREATED_ROLE_NAME}" --force --region "${AWS_REGION}"

# Get the cluster VPC ids
CLUSTER_VPC_IDS=$(aws eks describe-cluster --name "${CLUSTER_NAME}" --output json --region "${AWS_REGION}" \
    | jq -r '.cluster.resourcesVpcConfig.securityGroupIds[]')

# AWS describe cluster get vpc id
CLUSTER_VPC_ID=$(aws eks describe-cluster --name "${CLUSTER_NAME}" --output json --region "${AWS_REGION}" | jq -r '.cluster.resourcesVpcConfig.vpcId')

# Get the cluster public subnet ids as an array
# shellcheck disable=SC2016
CLUSTER_SUBNET_IDS=$(aws ec2 describe-subnets --filter "Name=vpc-id,Values=${CLUSTER_VPC_ID}" --query 'Subnets[?MapPublicIpOnLaunch==`true`].SubnetId' --output json --region "${AWS_REGION}" | jq -r '.[]')
# aws cli Create a DB subnet group
aws rds create-db-subnet-group --db-subnet-group-name "${DB_SUBNET_GROUP_NAME}" --db-subnet-group-description "DB subnet group - public subnets" --subnet-ids "$CLUSTER_SUBNET_IDS" --tags "Key=Name,Value=${GLOBAL_TAG}" --region "${AWS_REGION}"

# Amazon aws cli expose postgresql port 5432 on VPC security group
aws ec2 authorize-security-group-ingress --group-id "${CLUSTER_VPC_IDS}" --protocol tcp --port 5432 --cidr 0.0.0.0/0 --region "${AWS_REGION}"

cat <<EOF > gp3-def-sc.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: gp3
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
allowVolumeExpansion: true
provisioner: ebs.csi.aws.com
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: gp3
EOF

kubectl apply -f gp3-def-sc.yaml

if [ "$EMPTY_CLUSTER" == "true" ]; then
    echo "created empty cluster"
    echo "create database ${POSTGRES_SQL_ID}"
    echo "cluster VPC: ${CLUSTER_VPC_IDS}"
    echo "DB subnet group: ${DB_SUBNET_GROUP_NAME}"
    echo "Address: aws rds describe-db-instances --db-instance-identifier ${POSTGRES_SQL_ID} --output json --region \"${AWS_REGION}\" | jq -r '.DBInstances[0].Endpoint.Address'"
    exit
fi

# AWS CLI Create postgresql rds instance
aws rds create-db-instance \
    --db-instance-identifier "${POSTGRES_SQL_ID}" \
    --db-name "${POSTGRES_SQL_DB_NAME_1}" \
    --db-instance-class db.m6g.large \
    --engine postgres \
    --master-username "${POSTGRES_SQL_DB_USER_NAME}" \
    --master-user-password "${POSTGRES_SQL_DB_USER_PASSWORD}" \
    --storage-type io1 \
    --iops 1500 \
    --allocated-storage 100 \
    --vpc-security-group-ids "${CLUSTER_VPC_IDS}" \
    --db-subnet-group-name "${DB_SUBNET_GROUP_NAME}"  \
    --publicly-accessible \
    --tags "Key=Name,Value=${GLOBAL_TAG}" \
    --region "${AWS_REGION}" \
    --no-multi-az

# Check if the postgresql rds instance is available
aws rds wait db-instance-available --db-instance-identifier ${POSTGRES_SQL_ID} --region "${AWS_REGION}"

# Get the postgresql rds instance endpoint
POSTGRES_SQL_ENDPOINT=$(aws rds describe-db-instances --db-instance-identifier ${POSTGRES_SQL_ID} --output json --region "${AWS_REGION}" | jq -r '.DBInstances[0].Endpoint.Address')

# create a second database named "dex" in your RDS instance for Pachyderm's authentication service. 
# database must be named dex
# create a new user account and grant it full CRUD permissions to both pachyderm and dex databases.
PGPASSWORD="$POSTGRES_SQL_DB_USER_PASSWORD" psql -h "${POSTGRES_SQL_ENDPOINT}" -U "$POSTGRES_SQL_DB_USER_NAME" "$POSTGRES_SQL_DB_NAME_1" << EOF
SELECT datname FROM pg_database;
CREATE DATABASE dex;
CREATE USER $POSTGRES_SQL_DB_APP_USER_NAME WITH PASSWORD '${POSTGRES_SQL_DB_USER_PASSWORD}';
GRANT ALL PRIVILEGES ON DATABASE dex TO $POSTGRES_SQL_DB_APP_USER_NAME;
GRANT ALL PRIVILEGES ON DATABASE pachyderm TO $POSTGRES_SQL_DB_APP_USER_NAME;

SELECT usename AS role_name,
 CASE
  WHEN usesuper AND usecreatedb THEN
    CAST('superuser, create database' AS pg_catalog.text)
  WHEN usesuper THEN
    CAST('superuser' AS pg_catalog.text)
  WHEN usecreatedb THEN
    CAST('create database' AS pg_catalog.text)
  ELSE
    CAST('' AS pg_catalog.text)
 END role_attributes
FROM pg_catalog.pg_user
ORDER BY role_name desc;

EOF

cat <<EOF > values.yaml
deployTarget: AMAZON
etcd:
  storageClass: gp3
pachd:
  storage:
    amazon:
      bucket: ${BUCKET_NAME}
      region: ${AWS_REGION}
  serviceAccount:
    additionalAnnotations:
      eks.amazonaws.com/role-arn: $ROLE_ARN
  worker:
    serviceAccount:
      additionalAnnotations:
        eks.amazonaws.com/role-arn: $ROLE_ARN
  externalService:
    enabled: true
global:
  postgresql:
    postgresqlUsername: "${POSTGRES_SQL_DB_APP_USER_NAME}"
    postgresqlPassword: "${POSTGRES_SQL_DB_USER_PASSWORD}" 
    postgresqlDatabase: "pachyderm"
    postgresqlHost: "${POSTGRES_SQL_ENDPOINT}"
    postgresqlPort: "5432"

postgresql:
  enabled: false
EOF

helm repo add pach https://helm.pachyderm.com
helm repo update
helm install pachyderm -f ./values.yaml pach/pachyderm

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

STATIC_IP_ADDR=$(kubectl get svc pachd-lb -o json | jq -r .status.loadBalancer.ingress[0].hostname)
echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite
pachctl config set active-context ${CLUSTER_NAME}
pachctl config get active-context

echo "All done, run pachctl port-forward to resolve traffic then pachctl version to verify"
