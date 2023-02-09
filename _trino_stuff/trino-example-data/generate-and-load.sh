#!/bin/bash

set -ex

pachctl delete repo tpch || true
pachctl create repo tpch

# Remove all schemas except "tpch"
aws s3 cp s3://presto-example/v2/example-metadata.json - \
  | jq '{ tpch: .tpch }'  \
  | pachctl put file tpch@master:/metadata.json

for f in $(
  pachctl get file tpch@master:/metadata.json \
    | jq -r '.tpch[].sources[]'
); do
  # WHen I try to load this straight into pachd with put file ... -f s3://...
  # I get 'blob (key "v2/orders-1.csv") (code=Unknown): MissingRegion: could not find region configuration'
  # which I don't understand.
  aws s3 cp "s3://presto-example/v2/${f}" - | pachctl put file "tpch@master:/${f}"
done
