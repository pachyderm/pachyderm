#!/bin/bash

set -ex

dirs=(
  "${HOME}/.cache"
  "${HOME}/.cache/pip"
  "${HOME}/cached-deps"
)

for dir in "${dirs[@]}"; do
  if [[ -d "${dir}" ]]; then
    # change ownership and permissions on directory, in case it's owned by root
    sudo chown -R ${USER}:${USER} "${dir}"
    chmod 777 -R "${dir}"
  else
    # create directory
    mkdir -p "${dir}"
  fi
done
