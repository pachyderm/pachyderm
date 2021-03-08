SHELL := /bin/bash # Use bash syntax

.PHONY: test kubeval-gcp

test:
	go test -v -race ./... -count 1

kubeval-gcp:
	helm template pachyderm -f examples/gcp-values.yaml | kubeval --strict

kubeval-aws:
	helm template pachyderm -f examples/aws-values.yaml | kubeval --strict