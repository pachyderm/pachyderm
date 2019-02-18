#!/bin/bash
# This script pushes docker images to microk8s so that they can be
# pulled/run by kubernetes pods

if [[ $# -ne 1 ]]; then
  echo "error: need the name of the docker image to push"
fi

command -v pv >/dev/null 2>&1 || { echo >&2 "Required command 'pv' not found. Run 'sudo apt-get install pv'."; exit 1; }

docker save "${1}" | pv | (
  microk8s.docker load
)
