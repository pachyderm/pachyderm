#!/bin/bash

if ! pachctl version; then
  echo "Can't connect to cluster" >/dev/stderr
  exit 1
fi

if [[ "${VIRTUAL_ENV}" != "venv" ]]; then
  echo "Need to activate venv. Run 'source ./venv/bin/activate'" >/dev/stderr
  exit 1
fi

python3 ./albert-spark-example.py |& tee spark.log
kubectl get logs --tail=-1 -l app=pachd |& tee pachd.log
