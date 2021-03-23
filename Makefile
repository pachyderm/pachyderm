SHELL := /bin/bash # Use bash syntax


.PHONY: all test lint kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-local kubeval-minio kubeval-microsoft

all: pachyderm/values.schema.json

lint:
	helm lint pachyderm

test:  kubeval-aws kubeval-gcp kubeval-gcp-tls kubeval-local kubeval-minio kubeval-microsoft
	go test -race ./... -count 1

kubeval-aws:
	helm template pachyderm -f examples/aws-values.yaml | kubeval --strict

kubeval-gcp:
	helm template pachyderm -f examples/gcp-values.yaml | kubeval --strict

kubeval-gcp-tls:
	helm template pachyderm -f examples/gcp-values-tls.yaml | kubeval --strict

kubeval-local:
	helm template pachyderm -f examples/local-values.yaml | kubeval --strict

kubeval-minio:
	helm template pachyderm -f examples/minio-values.yaml | kubeval --strict

kubeval-microsoft:
	helm template pachyderm -f examples/microsoft-values.yaml | kubeval --strict

pachyderm/values.schema.json: pachyderm/values.yaml
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
