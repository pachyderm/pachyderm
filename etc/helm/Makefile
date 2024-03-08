# SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash -o pipefail # Use bash syntax


.PHONY: all test lint kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-local kubeval-minio kubeval-microsoft schema

all: pachyderm/values.schema.json

test: pachyderm/values.schema.json kubeval-aws-gp2 kubeval-aws-gp3 kubeval-gcp kubeval-gcp-tls kubeval-hub kubeval-local kubeval-local-dev kubeval-minio kubeval-microsoft

schema:
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
