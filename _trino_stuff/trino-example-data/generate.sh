#!/bin/bash

aws s3 cp s3://presto-example/v2/example-metadata.json example-metadata.json

# Remove all schemas except "tpch"
jq '{ tpch: .tpch }' example-metadata.json | sponge example-metadata-filtered.json

for t in $( jq -r '.tpch[].sources[]' example-metadata-filtered.json ); do
  aws s3 cp "s3://presto-example/v2/$t" "$t"
done

