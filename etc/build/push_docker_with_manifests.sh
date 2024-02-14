#!/bin/bash

set -euxo pipefail

# $VERSION is the tag to identify the manifest.  $UNDERLYING is the tag applied to the underlying
# containers.  For example, if you build images that are tagged pachyderm/pachd:SOMESHA1 and want
# that to become pachyderm/pachd:12.34.56, $VERSION should be 12.34.56 and $UNDERLYING should be
# SOMESHA1.

REPO=pachyderm
UNDERLYING=${UNDERLYING:-$VERSION}

for product in pachd worker pachctl; do
    echo ""
    echo "push $product tag=$VERSION (based on tag=$UNDERLYING)..."
    echo ""
    docker push "$REPO/$product-amd64:$UNDERLYING" || docker pull "$REPO/$product-amd64:$UNDERLYING"
    docker push "$REPO/$product-arm64:$UNDERLYING" || docker pull "$REPO/$product-arm64:$UNDERLYING"

    docker manifest create "$REPO/$product:$VERSION" \
           "$REPO/$product-amd64:$UNDERLYING" \
           "$REPO/$product-arm64:$UNDERLYING"

    docker manifest annotate "$REPO/$product:$VERSION" "$REPO/$product-amd64:$UNDERLYING" --arch amd64
    docker manifest annotate "$REPO/$product:$VERSION" "$REPO/$product-arm64:$UNDERLYING" --arch arm64

    docker manifest inspect "$REPO/$product:$VERSION"
    docker manifest push "$REPO/$product:$VERSION"
done
