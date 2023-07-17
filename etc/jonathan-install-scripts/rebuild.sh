#!/usr/bin/env bash
set -xeuo pipefail -o errexit

if [ '!' -d etc/helm/pachyderm ]; then
    echo "Run from a pachyderm checkout"
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

export VERSION=$(date +%s)
export PACHD_VERSION=localhost:5001/pachd:$VERSION
export WORKER_VERSION=localhost:5001/worker:$VERSION

GO="${GO:=go}"
COVERFLAGS="${COVERFLAGS:=}"

CGO_ENABLED=0 $GO build $COVERFLAGS -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./src/server/cmd/pachd
CGO_ENABLED=0 $GO build $COVERFLAGS -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./src/server/cmd/worker
CGO_ENABLED=0 $GO build -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./src/server/cmd/pachctl
CGO_ENABLED=0 $GO build -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./src/server/cmd/pachtf
CGO_ENABLED=0 $GO build -o worker_init -ldflags "-X github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=+$VERSION" ./etc/worker

docker build . -f ./Dockerfile.pachd -t $PACHD_VERSION
docker build . -f ./Dockerfile.worker -t $WORKER_VERSION

docker tag $PACHD_VERSION pachyderm/pachd:local
docker tag $WORKER_VERSION pachyderm/worker:local

docker push $PACHD_VERSION
docker push $WORKER_VERSION

cat > $SCRIPT_DIR/version.yaml <<EOF
pachd:
    image:
        repository: localhost:5001/pachd
        tag: "$VERSION"
    worker:
        image:
            repository: localhost:5001/worker
            tag: "$VERSION"
EOF

$SCRIPT_DIR/upgrade.sh

# cat <<EOF | kubectl patch deployment pachd --patch-file=/dev/stdin
# spec:
#     template:
#         spec:
#             containers:
#                 - name: pachd
#                   image: $PACHD_VERSION
#                   env:
#                       - name: WORKER_IMAGE
#                         value: $WORKER_VERSION
#                       - name: WORKER_SIDECAR_IMAGE
#                         value: $PACHD_VERSION
# EOF


kubectl rollout status deployment pachd --timeout=120s

set +x
kubectl get pod -o wide
echo "Waiting for new version to be serving."
for i in `seq 1 30`; do
    sleep 2
    date
    if pachctl version | grep $VERSION; then
        echo "New version is serving traffic."
        break
    fi
done
