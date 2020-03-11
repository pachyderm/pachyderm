#!/bin/bash

while [[ $# -gt 0 ]]
do
  key="$1"

  case $key in
      -i|--issuer)
      ISSUER="$2"
      shift
      shift
      ;;
      -j|--jwks-uri)
      JWKS_URI="$2"
      shift
      shift
      ;;
      -c|--client-id)
      CLIENT_ID="$2"
      shift
      shift
      ;;
      -h|--help)
      echo "
        Use Arguments:
          -i|--issuer Issuer for dex
          -j|--jwks-uri JWKS Key path provided by dex server
          -c|--client_id Client ID of the Dex Client set
      "
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

if [ -z "$ISSUER" ] || [ -z "$JWKS_URI" ] || [ -z "$CLIENT_ID" ];
then
  echo "
    Missing one of the options mentioned below:
    -i|--issuer Issuer for dex
    -j|--jwks-uri JWKS Key path provided by dex server
    -c|--client_id Client ID of the Dex Client set
  "
  exit
fi

cat << EOF > authentication/Istio/base/params.env
issuer=$ISSUER
client_id=$CLIENT_ID
jwks_uri=$JWKS_URI
EOF

kubectl create -f authorization/Kubernetes
kubectl create -f authorization/Istio
kustomize build authentication/Istio/base | kubectl apply -f -
