#!/bin/bash

set -euxo pipefail

tar -xvzf /tmp/workspace/dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp

sudo mv /tmp/pachctl_*/pachctl /usr/local/bin && chmod +x /usr/local/bin/pachctl

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

echo "{\"pachd_address\": \"grpcs://qa1.workspace.pachyderm.com:443\", \"session_token\": \"test\"}" | tr -d \\ | pachctl config set context "test" --overwrite
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

pushd examples/shuffle
    pachctl create repo fruits
    pachctl create repo pricing
    pachctl create pipeline -f shuffle.json
    pachctl put file fruits@master -f mango.jpeg
    pachctl put file fruits@master -f apple.jpeg
    pachctl put file pricing@master -f mango.json
    pachctl put file pricing@master -f apple.json

    # wait for everything to finish
    pachctl wait commit "shuffle@master"

    # check downloaded and uploaded bytes
    downloaded_bytes=$(pachctl list job -p shuffle --raw | jq '.stats.download_bytes | values')
    if [ "$downloaded_bytes" != "" ]; then
        echo "Unexpected downloaded bytes in shuffle repo: $downloaded_bytes"
        exit 1
    fi

    uploaded_bytes=$(pachctl list job -p shuffle --raw | jq '.stats.upload_bytes | values')
    if [ "$uploaded_bytes" != "" ]; then
        echo "Unexpected downloaded bytes in shuffle repo: $uploaded_bytes"
        exit 1
    fi

    # check that the files were made
    files=$(pachctl glob file "shuffle@master:**" --raw | jq '.file.path' -r)
    expected_files=$(echo -e "/apple/\n/apple/cost.json\n/apple/img.jpeg\n/mango/\n/mango/cost.json\n/mango/img.jpeg")
    if [ "$files" != "$expected_files" ]; then
        echo "Unexpected output files in shuffle repo: $files"
        exit 1
    fi
popd

pachctl delete pipeline --all
pachctl delete repo --all

pushd examples/word_count/
    pachctl create repo urls

    pushd data
        pachctl put file urls@master -f Wikipedia
    popd

    pushd pipelines
        pachctl create pipeline -f scraper.json
        pachctl create pipeline -f map.json
        pachctl create pipeline -f reduce.json
    popd

    pachctl wait commit "reduce@master"

    # just make sure we outputted some files that were right
    license_word_count=$(pachctl get file reduce@master:/license)
    if [ "$license_word_count" -lt 1 ]; then
        echo "Unexpected word count in reduce repo"
        exit 1
    fi
popd