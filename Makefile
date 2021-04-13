# SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash -o pipefail # Use bash syntax


.PHONY: all test lint kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-local kubeval-minio kubeval-microsoft

all: pachyderm/values.schema.json

lint:
	helm lint pachyderm

test: pachyderm/values.schema.json kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-hub kubeval-local kubeval-local-dev kubeval-minio kubeval-microsoft
	go test -race ./... -count 1

kubeval-aws:
	helm template pachyderm -f examples/aws-values.yaml | kubeval --strict

kubeval-gcp:
	helm template pachyderm -f examples/gcp-values.yaml | kubeval --strict

kubeval-gcp-tls:
	helm template pachyderm -f examples/gcp-values-tls.yaml | kubeval --strict

kubeval-hub:
	helm template pachyderm -f examples/hub-values.yaml | kubeval --strict

kubeval-local:
	helm template pachyderm -f examples/local-values.yaml | kubeval --strict

kubeval-local-dev:
	helm template pachyderm -f examples/local-dev-values.yaml | kubeval --strict

kubeval-minio:
	helm template pachyderm -f examples/minio-values.yaml | kubeval --strict

kubeval-microsoft:
	helm template pachyderm -f examples/microsoft-values.yaml | kubeval --strict

pachyderm/values.schema.json: pachyderm/values.yaml
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
