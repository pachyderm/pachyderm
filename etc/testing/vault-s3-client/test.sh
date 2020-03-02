#!/bin/bash

SCRIPT_DIR="$(dirname "${0}")"

set -ex

# Build pachyderm and push the relevant images
make install && make docker-build && make push-images || exit 1

# Start kops cluster
"${SCRIPT_DIR}/../deploy/aws.sh" --create --no-pachyderm

# Deploy vault
kubectl create -f -<<EOF
kind: Pod
apiVersion: v1
metadata:
  name: vault
  labels:
    app: vault
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
        - -dev-listen-address=0.0.0.0:8200
        - -log-level=debug
---
# service (so pachd can talk to vault)
kind: Service
apiVersion: v1
metadata:
  name: vault-svc
spec:
  selector:
    app: vault
  type: NodePort
  ports:
    - name: main
      port: 8200
EOF

# Wait for vault pod to come up
while [[ "$(kubectl get po/vault -o json | jq -r ".status.phase")" != "Running" ]]
do
  sleep 1
done

# port-forward to vault and connect vault client
kubectl port-forward vault 8200 &
sleep 3
export VAULT_ADDR='http://127.0.0.1:8200'
echo root | vault login -

# Activate and configure the vault s3 secret engine
vault secrets enable aws

AWS_ID=$(grep aws_access_key_id < ~/.aws/credentials | cut -d " " -f 3)
AWS_SECRET=$(grep aws_secret_access_key < ~/.aws/credentials | cut -d " " -f 3)
AWS_REGION=us-west-1  # default region in aws.sh
vault write aws/config/root \
  access_key="${AWS_ID}" \
  secret_key="${AWS_SECRET}" \
  region="${AWS_REGION}"

# Create S3 bucket
export STORAGE_SIZE=100
export BUCKET=${RANDOM}-pachyderm-store
aws s3api create-bucket \
  --bucket ${BUCKET} \
  --region ${AWS_REGION} \
  --create-bucket-configuration LocationConstraint=${AWS_REGION}
echo "BUCKET is ${BUCKET}"

# Create AWS IAM policy in vault, that allows anyone with access
# to this vault path to access Pachyderm's s3 bucket
VAULT_ROLE=pachd-object-store-role
vault write aws/roles/${VAULT_ROLE} policy=-<<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [ "s3:PutObject"
                , "s3:GetObject"
                , "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::${BUCKET}/*"
    },
    {
      "Effect": "Allow",
      "Action": [ "s3:ListBucket" ],
      "Resource": "arn:aws:s3:::${BUCKET}"
    }
  ]
}
EOF

# Create vault policy for accessing the above s3 path
curl -X POST -H "X-Vault-Token: root" --data '
{
    "policy": "{\"path\": {\"aws/creds/'"${VAULT_ROLE}"'\": { \"capabilities\": [\"read\"]}}}"
}' http://127.0.0.1:8200/v1/sys/policy/pachd-s3-policy

# Get vault token for Pachd
VAULT_TOKEN="$(
  curl -X POST -H "X-Vault-Token: root" --data '
  {
    "policies": ["pachd-s3-policy"]
  }' http://127.0.0.1:8200/v1/auth/token/create \
  | jq --raw-output ".auth.client_token"
)"

# Deploy
pachctl deploy amazon ${BUCKET} ${AWS_REGION} ${STORAGE_SIZE} \
  --dynamic-etcd-nodes=1 \
  --no-dashboard \
  --vault=http://vault-svc.default.svc.cluster.local:8200,"${VAULT_ROLE}","${VAULT_TOKEN}"
