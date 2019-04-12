#!/bin/bash

set -ex
# Runs various examples ot ensure they don't break. Some examples were
# designed for older versions of pachyderm and are not used here.

pushd examples/opencv
    pachctl --no-port-forwarding create repo images
    pachctl --no-port-forwarding create pipeline -f edges.json
    pachctl --no-port-forwarding create pipeline -f montage.json
    pachctl --no-port-forwarding put file images@master -i images.txt
    pachctl --no-port-forwarding put file images@master -i images2.txt

    # wait for everything to finish
    commit_id=`pachctl --no-port-forwarding list commit images -n 1 --raw | jq .commit.id -r`
    pachctl --no-port-forwarding flush job images@$commit_id

    # ensure the montage image was generated
    pachctl --no-port-forwarding inspect file montage@master:montage.png
popd

yes | pachctl --no-port-forwarding delete all

pushd examples/shuffle
    pachctl --no-port-forwarding create repo fruits
    pachctl --no-port-forwarding create repo pricing
    pachctl --no-port-forwarding create pipeline -f shuffle.json
    pachctl --no-port-forwarding put file fruits@master -f mango.jpeg
    pachctl --no-port-forwarding put file fruits@master -f apple.jpeg
    pachctl --no-port-forwarding put file pricing@master -f mango.json
    pachctl --no-port-forwarding put file pricing@master -f apple.json

    # wait for everything to finish
    commit_id=`pachctl --no-port-forwarding list commit fruits -n 1 --raw | jq .commit.id -r`
    pachctl --no-port-forwarding flush job fruits@$commit_id
    pachctl --no-port-forwarding flush commit fruits@$commit_id

    # check downloaded and uploaded bytes
    downloaded_bytes=`pachctl --no-port-forwarding list job -p shuffle --raw | jq '.stats.download_bytes | values'`
    if [ "$downloaded_bytes" != "" ]; then
        echo "Unexpected downloaded bytes in shuffle repo: $DOWNLOADED_BYTES"
        exit 1
    fi

    uploaded_bytes=`pachctl --no-port-forwarding list job -p shuffle --raw | jq '.stats.upload_bytes | values'`
    if [ "$uploaded_bytes" != "" ]; then
        echo "Unexpected downloaded bytes in shuffle repo: $uploaded_bytes"
        exit 1
    fi

    # check that the files were made
    files=`pachctl --no-port-forwarding list file "shuffle@master:*" --raw | jq '.file.path' -r`
    expected_files=`echo -e "/apple\n/apple/cost.json\n/apple/img.jpeg\n/mango\n/mango/cost.json\n/mango/img.jpeg"`
    if [ "$files" != "$expected_files" ]; then
        echo "Unexpected output files in shuffle repo: $files"
        exit 1
    fi
popd

yes | pachctl --no-port-forwarding delete all

pushd examples/word_count
    # note: we do not test reducing because it's slower
    pachctl --no-port-forwarding create repo urls
    pachctl --no-port-forwarding put file urls@master -f Wikipedia
    pachctl --no-port-forwarding create pipeline -f scraper.json
    pachctl --no-port-forwarding create pipeline -f map.json

    # wait for everything to finish
    commit_id=`pachctl --no-port-forwarding list commit urls -n 1 --raw | jq .commit.id -r`
    pachctl --no-port-forwarding flush commit urls@$commit_id

    # just make sure the count for the word 'wikipedia' is a valid and
    # positive int, since the specific count may vary over time
    wikipedia_count=`pachctl --no-port-forwarding get file map@master:wikipedia`
    if [ $wikipedia_count -le 0 ]; then
        echo "Unexpected count for the word 'wikipedia': $wikipedia_count"
        exit 1
    fi
popd

yes | pachctl --no-port-forwarding delete all

pushd examples/ml/hyperparameter
    pachctl --no-port-forwarding create repo raw_data
    pachctl --no-port-forwarding create repo parameters
    pachctl --no-port-forwarding list repo

    pushd data
        pachctl --no-port-forwarding put file raw_data@master:iris.csv -f noisy_iris.csv

        pushd parameters
            pachctl --no-port-forwarding put file parameters@master -f c_parameters.txt --split line --target-file-datums 1 
            pachctl --no-port-forwarding put file parameters@master -f gamma_parameters.txt --split line --target-file-datums 1
        popd
    popd

    pachctl --no-port-forwarding create pipeline -f split.json 
    pachctl --no-port-forwarding create pipeline -f model.json
    pachctl --no-port-forwarding create pipeline -f test.json 
    pachctl --no-port-forwarding create pipeline -f select.json

    commit_id=`pachctl --no-port-forwarding list commit raw_data -n 1 --raw | jq .commit.id -r`
    pachctl --no-port-forwarding flush job raw_data@$commit_id

    # just make sure we outputted some files
    selected_file_count=`pachctl --no-port-forwarding list file select@master | wc -l`
    if [ $selected_file_count -le 2 ]; then
        echo "Expected some files to be outputted in the select repo"
        exit 1
    fi
popd

yes | pachctl --no-port-forwarding delete all

pushd examples/ml/iris
    pachctl --no-port-forwarding create repo training
    pachctl --no-port-forwarding create repo attributes

    pushd data
        pachctl --no-port-forwarding put file training@master -f iris.csv
    popd

    pachctl --no-port-forwarding create pipeline -f julia_train.json

    pushd data/test
        pachctl --no-port-forwarding put file attributes@master -r -f .
    popd

    pachctl --no-port-forwarding list file attributes@master
    pachctl --no-port-forwarding create pipeline -f julia_infer.json

    commit_id=`pachctl --no-port-forwarding list commit training -n 1 --raw | jq .commit.id -r`
    pachctl --no-port-forwarding flush job training@$commit_id

    # just make sure we outputted some files
    inference_file_count=`pachctl --no-port-forwarding list file inference@master | wc -l`
    if [ $inference_file_count -ne 3 ]; then
        echo "Unexpected file count in inference repo"
        exit 1
    fi
popd
