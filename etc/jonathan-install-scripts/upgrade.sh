#!/usr/bin/env bash
set -xeu -o pipefail -o errexit

if [ '!' -d etc/helm/pachyderm ]; then
    echo "Run from a pachyderm checkout"
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

helm $@ upgrade pachyderm etc/helm/pachyderm -f $SCRIPT_DIR/hostname-values.yaml -f $SCRIPT_DIR/enterprise-key-values.yaml -f $SCRIPT_DIR/values.yaml -f $SCRIPT_DIR/version.yaml
