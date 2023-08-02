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

wget --directory-prefix=./data/ https://www.researchgate.net/profile/Pekka-Malo/publication/251231364_FinancialPhraseBank-v10/data/0c96051eee4fb1d56e000000/FinancialPhraseBank-v10.zip
unzip -o data/FinancialPhraseBank-v10.zip

# Upload the Financial Phrase Bank data
pachctl create repo financial_phrase_bank
pachctl put file financial_phrase_bank@master:/Sentences_50Agree.txt -f data/FinancialPhraseBank-v1.0/Sentences_50Agree.txt

# Download the pre-trained BERT language model
./download_model.sh

# Upload the language model to Pachyderm
pachctl create repo language_model
cd models/finbertTCR2/; pachctl put file -r language_model@master -f ./; cd ../../

# Set up labeled_data repo for labeling production data later
pachctl create repo labeled_data
pachctl start commit labeled_data@master; pachctl finish commit labeled_data@master

# Deploy the dataset creation pipeline
pachctl create pipeline -f pachyderm/dataset.json

# Deploy the training pipeline
pachctl create pipeline -f pachyderm/train_model.json

# (Optional) Use a sentiment word list to visualize the current dataset
pachctl create repo sentiment_words
pachctl put file sentiment_words@master:/LoughranMcDonald_SentimentWordLists_2018.csv -f resources/LoughranMcDonald_SentimentWordLists_2018.csv
pachctl create pipeline -f pachyderm/visualizations.json

# Tag our current dataset as branch v1 (easy to referr to later on)
pachctl create branch financial_phrase_bank@v1 --head master

# Modify the version of Financial Phrase Bank dataset used
pachctl start commit financial_phrase_bank@master
pachctl delete file financial_phrase_bank@master:/Sentences_50Agree.txt
pachctl put file financial_phrase_bank@master:/Sentences_AllAgree.txt -f data/FinancialPhraseBank-v1.0/Sentences_AllAgree.txt
pachctl finish commit financial_phrase_bank@master

# Version our new dataset
pachctl create branch financial_phrase_bank@v2 --head master

# Show the diff between the datasets
pachctl diff file financial_phrase_bank@v2 financial_phrase_bank@v1

# Inspect everything impacted by v1 of our dataset
pachctl wait commit financial_phrase_bank@v1 --raw

# Download the trained model
pachctl get file -r train_model@master:/ -o trained_model/

pachctl delete pipeline --all
pachctl delete repo --all

