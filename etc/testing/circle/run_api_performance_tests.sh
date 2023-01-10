#!/bin/bash

set -euxo pipefail

pachctl_tag=$PACHD_VERSION 
if [ "$PACHD_VERSION" == "latest" ]
then 
    if [ -z "$PACHD_LATEST_VERSION" ]
    then 
        pachctl_tag=$(git tag --sort=taggerdate | tail -1) # the latest tag should be the nightly
    else
        pachctl_tag="$PACHD_LATEST_VERSION"
    fi
fi
image_tag="${pachctl_tag//v/}" #removing v from the front for the image

# Install pachctl
gh release download "$pachctl_tag" --pattern "pachctl_${image_tag}_amd64.deb" --repo pachyderm/pachyderm --output /tmp/pachctl.deb
sudo dpkg -i /tmp/pachctl.deb

kubectl config use-context minikube
# Create Pachyderm deployment
helm install pachyderm etc/helm/pachyderm \
    -f etc/testing/circle/helm-values.yaml \
    --set pachd.image.tag="$image_tag"
kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

pachctl config import-kube perftest
echo "$ENT_ACT_CODE" | pachctl license activate
pachctl version
nohup pachctl port-forward &

# Run locust tests
cd locust-pachyderm
pip3 install -r requirements.txt
locust --version
locust -f locustfile.py --headless --users 50 --spawn-rate 1 --run-time 3m \
    --csv /tmp/test-results/api-perf \
    --html /tmp/test-results/api-perf-stats.html \
    --exit-code-on-error 0 # errors are reported in the artifacts, if the test finishes the pipeline ran successfully
cd ..

# Collect and export stats to bigquery
export PACHD_PERF_VERSION="$image_tag"
export API_PERF_RESULTS_FOLDER="/tmp/test-results"
cd etc/testing/circle/workloads/api-perf-collector
pip3 install -r requirements.txt
python3 app.py
cd ..
