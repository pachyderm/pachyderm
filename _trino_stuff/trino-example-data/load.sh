#!/bin/bash

pachctl delete repo tpch || true
pachctl create repo tpch

pachctl put file tpch@master:/metadata.json -f example-metadata-filtered.json
for f in _trino_stuff/trino-example-data/*.csv; do
  pachctl put file tpch@master:/$(basename $f) -f "$f"
done
