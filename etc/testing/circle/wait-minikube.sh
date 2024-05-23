#!/bin/bash

set -Eex

# If ts command is not present.
maybe_ts() {
  if command -v ts ; then
    ts
  else
    cat
  fi
}

# Try to connect for three minutes
for _ in $(seq 36); do
    if kubectl version &>/dev/null; then
        echo 'minikube ready' | maybe_ts
        exit 0
    fi
    echo 'sleeping' | maybe_ts
    sleep 5
done

# Give up--kubernetes isn't coming up
exit 1
