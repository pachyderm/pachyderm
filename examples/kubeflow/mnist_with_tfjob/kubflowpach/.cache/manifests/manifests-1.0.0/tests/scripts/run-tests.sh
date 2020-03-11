#!/bin/bash

# Copyright 2019 The Kubeflow Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This shell script is used to build an image from our argo workflow

set -o errexit
set -o nounset
set -o pipefail

REPO="$(pwd)"

cd tests

# the tests depend on kustomize
export PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}
export GO111MODULE=on
make test

# Python test for the kustomization images
# TODO(yanniszark): install these in the worker image
# TODO(https://github.com/kubeflow/manifests/issues/449): 
# The code below doesn't properly handle the case where digest 
# is used.
#pip install -r "${REPO}/tests/scripts/requirements.txt"
#python3 "${REPO}"/tests/scripts/extract_images.py "${REPO}"

#if [[ `git status --porcelain` ]]; then
#    echo "Error: images missing from kustomization files."
#    git --no-pager diff
#    exit 1
#fi