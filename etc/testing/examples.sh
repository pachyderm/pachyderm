#!/bin/bash

set -ex

if [[ -n $CIRCLE_SHA1 ]]; then
    helm install pachyderm etc/helm/pachyderm -f etc/testing/circle/helm-values.yaml --set pachd.image.tag="${CIRCLE_SHA1}"
fi

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

for i in $(seq 1 20); do
    if pachctl version; then
        echo "pachd ready after $i attempts"
        break
    else
        sleep 5
        continue
    fi
done

# Runs various examples to ensure they don't break. Some examples were
# designed for older versions of pachyderm and are not used here.

# NOTE: this script is run periodically in hub as a coarse-grained end-to-end
# test. Be careful to ensure changes here work fine on hub. See hub's
# examples-runner for details.

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

## TODO(2.0 optional): Add wikipedia file to CI? Also, this test will need to be rewritten
## when the file part change PR is merged.
##pushd examples/word_count
##    # note: we do not test reducing because it's slower
##    pachctl create repo urls
##    pachctl put file urls@master -f Wikipedia
##    pachctl create pipeline -f scraper.json
##    pachctl create pipeline -f map/map.json
##
##    # wait for everything to finish
##    commit_id=$(pachctl list commit urls -n 1 --raw | jq .commit.id -r)
##    pachctl flush commit "urls@$commit_id"
##
##    # just make sure the count for the word 'wikipedia' is a valid and
##    # positive int, since the specific count may vary over time
##    wikipedia_count=$(pachctl get file map@master:wikipedia)
##    if [ "$wikipedia_count" -le 0 ]; then
##        echo "Unexpected count for the word 'wikipedia': $wikipedia_count"
##        exit 1
##    fi
##popd
##
##pachctl delete pipeline --all
##pachctl delete repo --all

## TODO(2.0 optional): Implement put file split.
##pushd examples/ml/hyperparameter
##    pachctl create repo raw_data
##    pachctl create repo parameters
##    pachctl list repo
##
##    pushd data
##        pachctl put file raw_data@master:iris.csv -f noisy_iris.csv
##
##        pushd parameters
##            pachctl put file parameters@master -f c_parameters.txt --split line --target-file-datums 1
##            pachctl put file parameters@master -f gamma_parameters.txt --split line --target-file-datums 1
##        popd
##    popd
##
##    pachctl create pipeline -f split.json
##    pachctl create pipeline -f model.json
##    pachctl create pipeline -f test.json
##    pachctl create pipeline -f select.json
##
##    pachctl wait commit "select@master"
##
##    # just make sure we outputted some files
##    selected_file_count=$(pachctl list file select@master | wc -l)
##    if [ "$selected_file_count" -le 2 ]; then
##        echo "Expected some files to be outputted in the select repo"
##        exit 1
##    fi
##popd
##
##pachctl delete pipeline --all
##pachctl delete repo --all

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
