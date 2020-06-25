#!/bin/bash

set -ex

# Runs various examples to ensure they don't break. Some examples were
# designed for older versions of pachyderm and are not used here.

function delete_all {
    # Rather than calling `pachctl delete all`, which also deletes auth
    # credentials (if any), these commands will only delete the bits that are
    # relevant to the examples
    pachctl delete pipeline --all
    pachctl delete repo --all
}

pushd examples/opencv
    pachctl create repo images
    pachctl create pipeline -f edges.json
    pachctl create pipeline -f montage.json
    pachctl put file images@master -i images.txt
    pachctl put file images@master -i images2.txt

    # wait for everything to finish
    commit_id=$(pachctl list commit images -n 1 --raw | jq .commit.id -r)
    pachctl flush job "images@$commit_id"

    # ensure the montage image was generated
    pachctl inspect file montage@master:montage.png
popd

delete_all

pushd examples/shuffle
    pachctl create repo fruits
    pachctl create repo pricing
    pachctl create pipeline -f shuffle.json
    pachctl put file fruits@master -f mango.jpeg
    pachctl put file fruits@master -f apple.jpeg
    pachctl put file pricing@master -f mango.json
    pachctl put file pricing@master -f apple.json

    # wait for everything to finish
    commit_id=$(pachctl list commit fruits -n 1 --raw | jq .commit.id -r)
    pachctl flush job "fruits@$commit_id"
    pachctl flush commit "fruits@$commit_id"

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
    files=$(pachctl list file "shuffle@master:*" --raw | jq '.file.path' -r)
    expected_files=$(echo -e "/apple\n/apple/cost.json\n/apple/img.jpeg\n/mango\n/mango/cost.json\n/mango/img.jpeg")
    if [ "$files" != "$expected_files" ]; then
        echo "Unexpected output files in shuffle repo: $files"
        exit 1
    fi
popd

delete_all

pushd examples/word_count
    # note: we do not test reducing because it's slower
    pachctl create repo urls
    pachctl put file urls@master -f Wikipedia
    pachctl create pipeline -f scraper.json
    pachctl create pipeline -f map.json

    # wait for everything to finish
    commit_id=$(pachctl list commit urls -n 1 --raw | jq .commit.id -r)
    pachctl flush commit "urls@$commit_id"

    # just make sure the count for the word 'wikipedia' is a valid and
    # positive int, since the specific count may vary over time
    wikipedia_count=$(pachctl get file map@master:wikipedia)
    if [ "$wikipedia_count" -le 0 ]; then
        echo "Unexpected count for the word 'wikipedia': $wikipedia_count"
        exit 1
    fi
popd

delete_all

pushd examples/ml/hyperparameter
    pachctl create repo raw_data
    pachctl create repo parameters
    pachctl list repo

    pushd data
        pachctl put file raw_data@master:iris.csv -f noisy_iris.csv

        pushd parameters
            pachctl put file parameters@master -f c_parameters.txt --split line --target-file-datums 1 
            pachctl put file parameters@master -f gamma_parameters.txt --split line --target-file-datums 1
        popd
    popd

    pachctl create pipeline -f split.json 
    pachctl create pipeline -f model.json
    pachctl create pipeline -f test.json 
    pachctl create pipeline -f select.json

    commit_id=$(pachctl list commit raw_data -n 1 --raw | jq .commit.id -r)
    pachctl flush job "raw_data@$commit_id"

    # just make sure we outputted some files
    selected_file_count=$(pachctl list file select@master | wc -l)
    if [ "$selected_file_count" -le 2 ]; then
        echo "Expected some files to be outputted in the select repo"
        exit 1
    fi
popd

delete_all

pushd examples/ml/iris
    pachctl create repo training
    pachctl create repo attributes

    pushd data
        pachctl put file training@master -f iris.csv
    popd

    pachctl create pipeline -f julia_train.json

    pushd data/test
        pachctl put file attributes@master -r -f .
    popd

    pachctl list file attributes@master
    pachctl create pipeline -f julia_infer.json

    commit_id=$(pachctl list commit training -n 1 --raw | jq .commit.id -r)
    pachctl flush job "training@$commit_id"

    # just make sure we outputted some files
    inference_file_count=$(pachctl list file inference@master | wc -l)
    if [ "$inference_file_count" -ne 3 ]; then
        echo "Unexpected file count in inference repo"
        exit 1
    fi
popd
