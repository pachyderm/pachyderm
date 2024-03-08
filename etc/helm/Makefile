# SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash -o pipefail # Use bash syntax


.PHONY: all schema

all: pachyderm/values.schema.json

schema:
	helm schema-gen pachyderm/values.yaml > pachyderm/values.schema.json
