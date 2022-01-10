#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

# Deploy Loki but don't block until it's deployed - we'll check later
helm repo remove grafana || true
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install loki grafana/loki-stack
