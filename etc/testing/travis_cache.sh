#!/bin/bash

# Create all of our Travis cached dirs, or, if they exist, loosen their
# permissions enough that our build process is guaranteed to be able to write
# to them and Travis's caching process is guaranteed to be able to read them.
# (Travis caching requires that all cached files be readable, and, I surmise,
# that cached directories be executable, to be traversed).

set -ex

# If any directories are added or changed here, they must also be added to
# .travis.yml
dirs=(
  "${HOME}/.cache"
  "${HOME}/cached-deps"
  "${GOPATH}/pkg"
  "$(python3 -c 'import site; print(site.USER_BASE)')" # $HOME/.local
)

for dir in "${dirs[@]}"; do
  if [[ -d "${dir}" ]]; then
    # change ownership on directory, in case it's owned by root
    sudo chown -R "${USER}:${USER}" "${dir}"
  else
    # create directory
    mkdir -p "${dir}"
  fi
  # Loosen permissions so build processes can write and Travis can cache
  chmod 777 -R "${dir}"
done
