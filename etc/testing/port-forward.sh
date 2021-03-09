#!/bin/bash

set -euo pipefail

if [ -f /tmp/port-forwards/postgres.pid ]; then
  kill $(cat /tmp/port-forwards/postgres.pid) || true
fi

kubectl port-forward service/postgres 32228:5432 --address 127.0.0.1 >/dev/null 2>&1 &
PID=$!

mkdir -p /tmp/port-forwards
echo $PID > /tmp/port-forwards/postgres.pid
