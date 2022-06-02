#!/usr/bin/env bash
set -xeuo pipefail

export VERSION=localhost:5001/pachd:$(date +%s)
CGO_ENABLED=0 go build -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./src/server/cmd/pachd
docker build . -f ./Dockerfile.pachd -t $VERSION
docker tag $VERSION pachyderm/pachd:local
docker push $VERSION
kind load docker-image pachyderm/pachd:local
kubectl set image deployment pachd pachd=$VERSION
kubectl rollout status deployment pachd --timeout=10s

set +x
kubectl get pod -o wide
echo "Waiting for new version to be serving."
for i in `seq 1 30`; do
    sleep 2
    if pachctl version | grep $VERSION; then
        echo "New version is serving traffic."
        break
    fi
done
