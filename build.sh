#!/bin/bash
set -e

VERSION=$(date +%s)
go get github.com/go-bindata/go-bindata/...
go-bindata -o src/server/cmd/worker/assets/assets.go -pkg assets /etc/ssl/certs/...
CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o pachd "src/server/cmd/pachd/main.go"
CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o worker "src/server/cmd/worker/main.go"
docker build -f Dockerfile.pachd . -t localhost:5000/pachd:1.13-$VERSION
docker push localhost:5000/pachd:1.13-$VERSION
echo $VERSION
sudo cp pachd /pachd
kubectl set image deployment/pachd pachd=localhost:5000/pachd:1.13-$VERSION
kubectl rollout status deployment pachd
