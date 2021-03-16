SHELL := /bin/bash # Use bash syntax

.PHONY: test lint kubeval-gcp kubeval-aws

lint:
	helm lint pachyderm

test:
	go test -v -race ./... -count 1

kubeval-gcp:
	helm template pachyderm -f examples/gcp-values.yaml | kubeval --strict

kubeval-aws:
	helm template pachyderm -f examples/aws-values.yaml | kubeval --strict

pachyderm/values.schema.json: pachyderm/values.yaml
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
