SHELL := /bin/bash # Use bash syntax

.PHONY: test

test:
	go test -v -race ./... -count 1
