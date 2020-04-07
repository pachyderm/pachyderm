#!/bin/sh
# wait-for-jobs will wait until there are no pachyderm jobs either succeeded or succeeded/failed
# It takes 2 arguments.
# $(1) can be "succeeded-only" "succeeded-or-failed"
# $(2) description: some text that will inform the user what kind of jobs are done.

HERE="$(dirname "$0")"
# shellcheck source=./etc/examples/paths.sh
. "${HERE}/paths.sh"

# the jq search filter: the filter '.state!="JOB_SUCCESS"' (without single quotes) will wait until all jobs are successful
#                            the filter '.state!="JOB_SUCCESS" and .state!="JOB_FAILURE" will wait until all jobs have
#                               either succeeded or failed.
#                            You can put any jq filter that will work with the "--raw" output of "list job" to produce
#                               a list of job IDs.

if [ $# -eq 0 ]
then
    ${ECHO} "No arguments supplied."
    exit 1
elif [ $# -lt 2 ]
then
     ${ECHO} "Too few arguments supplied."
     exit 1
fi
 
if [ -z "$1" ] || [ -z "$2" ]
then
    ${ECHO} "No argument supplied."
    exit 1
fi

if [ "$1" = "succeeded-only" ]
then
    filter='.state!="JOB_SUCCESS"'
elif [ "$1" = "succeeded-or-failed" ] 
then
    filter='.state!="JOB_SUCCESS" and .state!="JOB_FAILURE"'
else
    ${ECHO} 'First argument must be "succeeded-only" or "succeeded-or-failed".'
    exit 1
fi

while
    JOBS=$(${PACHCTL} list job --raw | jq "select(${filter})|.job.id") && \
	NUMJOBS=$(${ECHO} -n "${JOBS}" | wc -w | tr -d " ") && \
	[ "${NUMJOBS}" -gt 0 ] 
do
    if      
        [ "${NUMJOBS}" -gt 1 ] 
    then
        STATUS_MSG="Waiting for ${NUMJOBS} jobs to finish..."
    else
        STATUS_MSG="Waiting for ${JOBS} job to finish..."
    fi
    ${ECHO}  -en "\e[G\e[K${STATUS_MSG}"
    sleep 1
done

${ECHO} -e "\e[G\e[K${2}"

