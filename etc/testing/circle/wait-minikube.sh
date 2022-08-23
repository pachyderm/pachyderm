#!/bin/bash

set -Eex

# Try to connect for three minutes
for _ in $(seq 36); do
    if kubectl version &>/dev/null; then
        echo 'minikube ready' | ts
        exit 0
    fi
    echo 'sleeping' | ts
    sleep 5
done

# Give up--kubernetes isn't coming up
exit 1
