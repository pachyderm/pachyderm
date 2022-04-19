#!/usr/bin/env bash
set -xeuo pipefail

export VERSION=localhost:5001/pachd:$(date +%s)
CGO_ENABLED=0 go build ./src/server/cmd/pachd
docker build . -f ./Dockerfile.pachd -t $VERSION
docker push $VERSION
kubectl set image deployment pachd pachd=$VERSION
kubectl rollout status deployment pachd --timeout=10s
kubectl get pod -o wide
