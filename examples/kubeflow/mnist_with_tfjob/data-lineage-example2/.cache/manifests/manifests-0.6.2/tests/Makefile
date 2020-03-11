# Copyright 2017 The Kubernetes Authors.
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
#
GOLANG_VERSION ?= 1.12.4
GOPATH ?= $(HOME)/go

# Comma seperated items within {} for more than one file
EXCLUDE ?= istio-install-base_test.go

export GO111MODULE = on
export GO = go

all: test

generate: clean
	@cd .. && hack/gen-test-targets.sh && [ -n "$(EXCLUDE)" ] && rm -f tests/$(EXCLUDE) || echo done
	$(GO) fmt ./...

test:
	@GO111MODULE=on $(GO) test -v .

clean:
	@rm -f $$(ls *_test.go | egrep -v 'kusttestharness_test.go|go.mod|go.sum')

run-unittest-plugin:
	cd .. && XDG_CONFIG_HOME=$$(pwd)/plugins kustomize --enable_alpha_goplugins_accept_panic_risk build profiles/overlays/test

run-application-plugin:
	cd .. && XDG_CONFIG_HOME=$$(pwd)/plugins kustomize --enable_alpha_goplugins_accept_panic_risk build jupyter/jupyter-web-app/overlays/application
