#!/bin/bash

set -euxo pipefail

tar -xvzf /tmp/workspace/dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp

sudo mv /tmp/pachctl_*/pachctl /usr/local/bin && chmod +x /usr/local/bin/pachctl

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

echo "{\"pachd_address\": \"grpc://${PACHD_IP}:80\", \"session_token\": \"test\"}" | tr -d \\ | pachctl config set context "test" --overwrite
pachctl config set active-context "test"

# Print client and server versions, for debugging.  (Also waits for proxy to discover pachd, etc.)
READY=false
for i in $(seq 1 20); do
    if pachctl version; then
        echo "pachd ready after $i attempts"
        READY=true
        break
    else
        sleep 5
        continue
    fi
done

if [ "$READY" = false ] ; then
    echo 'pachd failed to start'
    exit 1
fi

pushd examples/opencv
    pachctl create repo images
    pachctl create pipeline -f edges.json
    pachctl create pipeline -f montage.json
    pachctl put file images@master -i images.txt
    pachctl put file images@master -i images2.txt

    # wait for everything to finish
    pachctl wait commit "montage@master"

    # ensure the montage image was generated
    pachctl inspect file montage@master:montage.png
popd

pachctl delete pipeline --all
pachctl delete repo --all