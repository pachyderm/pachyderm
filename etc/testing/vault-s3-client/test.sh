#!/bin/bash

SCRIPT_DIR="$(dirname "${0}")"

# Start kops cluster
${SCRIPT_DIR}/../deploy/aws.sh --no-pachyderm

# Deploy vault
kubectl create -f -<<EOF
kind: Pod
apiVersion: v1
metadata:
  name: vault
spec:
  containers:
    - name: vault
      image: vault
      command:
        - vault
        - server
      args:
        - -dev
        - -dev-root-token-id=root
        - -log-level=debug
---
# service (so pachd can talk to vault)
kind: Service
apiVersion: v1
metadata:
  name: vault-svc
spec:
  selector:
    name: vault
  type: NodePort
  ports:
    - name: main
      port: 8200
EOF

# port-forward to vault and connect vault client
kubectl port-forward vault 8200
export VAULT_ADDR='http://127.0.0.1:8200'
echo root | vault login -

# Activate and configure the vault s3 secret engine
vault secrets enable aws

AWS_ID=`cat ~/.aws/credentials | grep aws_access_key_id  | cut -d " " -f 3`
AWS_SECRET=`cat ~/.aws/credentials | grep aws_secret_access_key | cut -d " " -f 3`
AWS_AVAILABILITY_ZONE=us-west-1a
AWS_REGION=us-west-1  # default region in aws.sh
vault write aws/config/root \
  access_key="${AWS_ID}" \
  secret_key="${AWS_SECRET}" \
  region="${AWS_REGION}"

# Create S3 bucket
export STORAGE_SIZE=100
export BUCKET_NAME=${RANDOM}-pachyderm-store
aws s3api create-bucket \
  --bucket s3://${BUCKET} \
  --region ${AWS_REGION} \
  --create-bucket-configuration LocationConstraint=${AWS_REGION}

# Create vault policy for accessing s3 bucket
vault write aws/roles/pachd-object-store-role policy=-<<EOF
{
  "Version": "2018-3-28",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [ "s3:PutObject"
                , "s3:GetObject"
                , "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::${BUCKET}/*"
    }
  ]
}
EOF

# Deploy
pachctl deploy amazon ${BUCKET_NAME} ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --no-dashboard --credentials=${AWS_ID},${AWS_KEY},

