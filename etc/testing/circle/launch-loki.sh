#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

# Deploy Loki but don't block until it's deployed - we'll check later
helm repo remove loki || true
helm repo add loki https://grafana.github.io/loki/charts
helm repo update
helm upgrade --install loki loki/loki-stack
