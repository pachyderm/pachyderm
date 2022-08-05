#!/bin/bash

set -euxo pipefail

REPO=pachyderm

for product in pachd worker pachctl mount-server; do
    echo "push $product $VERSION..."
    docker push "$REPO/$product-amd64:$VERSION"
    docker push "$REPO/$product-arm64:$VERSION"

    docker manifest create "$REPO/$product:$VERSION" \
           "$REPO/$product-amd64:$VERSION" \
           "$REPO/$product-arm64:$VERSION"

    docker manifest annotate "$REPO/$product:$VERSION" "$REPO/$product-amd64:$VERSION" --arch amd64
    docker manifest annotate "$REPO/$product:$VERSION" "$REPO/$product-arm64:$VERSION" --arch arm64

    docker manifest inspect "$REPO/$product:$VERSION"
    docker manifest push "$REPO/$product:$VERSION"
done
