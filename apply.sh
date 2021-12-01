#!/bin/sh

set -ve

jsonnet $1 | jq '.[] ' | pachctl update pipeline
