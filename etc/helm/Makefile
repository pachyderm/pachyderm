# SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash -o pipefail # Use bash syntax


.PHONY: all test lint kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-local kubeval-minio kubeval-microsoft schema

all: pachyderm/values.schema.json

lint:
	helm lint pachyderm --set deployTarget=LOCAL

test: pachyderm/values.schema.json kubeval-aws-gp2 kubeval-aws-gp3 kubeval-gcp kubeval-gcp-tls kubeval-hub kubeval-local kubeval-local-dev kubeval-minio kubeval-microsoft
	go test -race ./... -count 1

kubeval-aws-gp2:
	helm template pachyderm -f examples/aws-gp2-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-aws-gp3:
	helm template pachyderm -f examples/aws-gp3-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-gcp:
	helm template pachyderm -f examples/gcp-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-gcp-tls:
	helm template pachyderm -f examples/gcp-values-tls.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-hub:
	helm template pachyderm -f examples/hub-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-local-dev:
	helm template pachyderm -f examples/local-dev-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-minio:
	helm template pachyderm -f examples/minio-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-microsoft:
	helm template pachyderm -f examples/microsoft-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

kubeval-enterprise:
	helm template pachyderm -f examples/enterprise-values.yaml | kubeval --schema-location https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master --strict

schema:
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
