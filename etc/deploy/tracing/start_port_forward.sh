#!/bin/bash

# Re-establish port-forward to jaeger
jaeger_pod="$(kubectl get po -l app=jaeger -o jsonpath='{.items[].metadata.name}')"
nohup kubectl port-forward "po/${jaeger_pod}" 16686 &  # UI port
nohup kubectl port-forward "po/${jaeger_pod}" 14268 &  # Collector port
nohup kubectl port-forward "po/etcd-0" 2379 & # etcd port-forward

cat <<EOF
#####################
# Connect pachctl to Jaeger with:
export JAEGER_ENDPOINT=localhost:14268
#####################
EOF

