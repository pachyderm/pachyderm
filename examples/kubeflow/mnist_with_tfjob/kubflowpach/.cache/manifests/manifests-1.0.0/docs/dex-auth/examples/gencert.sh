#!/bin/bash

# TODO(krishnadurai): Remove this file as soon as cert tooling is introduced in kfctl
# Tracking issue: https://github.com/kubeflow/kfctl/issues/6

mkdir -p ssl

while [[ $# -gt 0 ]]
do
  key="$1"

  case $key in
      -d|--dex-domain)
      DEX_DOMAIN="$2"
      shift
      shift
      ;;
      -h|--help)
      echo "Use -d|--dex-domain to supply domain name for dex server"
      exit
      shift
      shift
      ;;
      *)    # unknown option
      echo "Invalid option -$2" >&2
      exit
      ;;
  esac
done

if [ -z "$DEX_DOMAIN" ];
then
  echo "Enter -d|--dex-domain to supply domain name for dex server"
  exit
fi

cat << EOF > ssl/req.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = dex.$DEX_DOMAIN
DNS.2 = login.$DEX_DOMAIN
DNS.3 = ldap-admin.$DEX_DOMAIN
EOF

openssl genrsa -out ssl/ca-key.pem 2048
openssl req -x509 -new -nodes -key ssl/ca-key.pem -days 1000 -out ssl/ca.pem -subj "/CN=kube-ca"

openssl genrsa -out ssl/key.pem 2048
openssl req -new -key ssl/key.pem -out ssl/csr.pem -subj "/CN=kube-ca" -config ssl/req.cnf
openssl x509 -req -in ssl/csr.pem -CA ssl/ca.pem -CAkey ssl/ca-key.pem -CAcreateserial -out ssl/cert.pem -days 1000 -extensions v3_req -extfile ssl/req.cnf
