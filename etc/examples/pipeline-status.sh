#!/bin/bash
# pipeline-status will return True if the pipeline specified as its argument is ready, False otherwise
# takes 1 argument, the name of the pipeline
# We are using the .items[*] syntax below, instead of .items[0],
# because items[0] will produce an array bounds in kubectl error if the pipeline doesn't exist yet.
# With .items[*], kubectl will return an empty string rather than error out.

HERE="$(dirname "$0")"
# shellcheck source=./etc/examples/paths.sh
. "${HERE}/paths.sh"

if [ $# -eq 0 ]
then
    ${ECHO} "No arguments supplied."
    exit 1
fi

if [ -z "$1" ]
then
    ${ECHO} "No argument supplied."
    exit 1
fi


CURRENT_STATUS=$(${KUBECTL} get pod -l suite=pachyderm,pipelineName="$1" \
			-o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}')

if [ "$CURRENT_STATUS" ]
then
    ${ECHO} -n "$CURRENT_STATUS"
else
    ${ECHO} -n False
fi

