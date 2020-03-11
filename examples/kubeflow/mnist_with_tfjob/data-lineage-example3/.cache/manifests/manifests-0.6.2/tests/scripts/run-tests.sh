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

# the tests depend on kustomize
export PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}
export GO111MODULE=on
go get sigs.k8s.io/kustomize

# Generate Go files on upstream HEAD 
cd /src/kubeflow/manifests/tests
make generate
cd -
if diff /src/kubeflow/manifests/tests/*.go tests/*.go ; then
    cd tests
    make test
else
    echo "Tests failed. Please run `make generate` and re-run tests"
    exit 1
fi
