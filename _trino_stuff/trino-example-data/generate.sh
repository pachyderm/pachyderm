#!/bin/bash

# Remove all schemas except "tpch"
aws s3 cp s3://presto-example/v2/example-metadata.json - \
  | jq '{ tpch: .tpch }' >metadata.json

for t in $( jq -r '.tpch[].sources[]' metadata.json ); do
  aws s3 cp "s3://presto-example/v2/$t" "$t"
done

