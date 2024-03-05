#!/bin/bash

set -exo pipefail

function get_test_idx {
    # Basically, this formula is modulo for round robin distribution. When i=0 it's just the current node index.
    # Each iteration after, we add the total node number to get the next value who's modulo would be the current node index.
    echo $(( "${CIRCLE_NODE_INDEX}" + ( "$1" * "${CIRCLE_NODE_TOTAL}" ) ))
}

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

helm repo add pachyderm https://helm.pachyderm.com

helm repo update

if [ "$CI" ] && [ "$CIRCLE_NODE_TOTAL" -gt 0 ]
then
    mapfile -t tests < <(go test -v -tags k8s ./src/testing/deploy/ -list=. | grep -v '^ok')
    total_tests="${#tests[@]}"
    test_names=""
    i=0
    test_idx=$(get_test_idx "$i")
    while [ "$test_idx" -lt "$total_tests" ]
    do 
        test_names+="^${tests[${test_idx}]}\$|"
        i=$((i+1))
        test_idx=$(get_test_idx "$i")
    done
    if [ "${#test_names}" -gt 0 ]
    then
        test_names="${test_names::-1}" # trim last |
       gotestsum --raw-command --debug --junitfile=/tmp/test-results/circle/gotestsum-report.xml -- go test -v=test2json -run "$test_names" -failfast -parallel=3 -timeout 3600s ./src/testing/deploy -tags=k8s -cover -test.gocoverdir="$TEST_RESULTS" -covermode=atomic -coverpkg=./... | stdbuf -i0 tee -a /tmp/go-test-results.txt
    fi
else
    go test -v=test2json -failfast -parallel=3 -timeout 3600s ./src/testing/deploy -tags=k8s -cover -test.gocoverdir="$TEST_RESULTS" -covermode=atomic -coverpkg=./... | stdbuf -i0 tee -a /tmp/go-test-results.txt
fi
