#!/bin/bash

set -Eex

# Try to connect for three minutes
for _ in $(seq 36); do
    if kubectl version &>/dev/null; then
        echo 'minikube ready'
        exit 0
    fi
    echo 'sleeping'
    sleep 5
done

# Give up--kubernetes isn't coming up
exit 1
