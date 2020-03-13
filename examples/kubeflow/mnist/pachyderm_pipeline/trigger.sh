#!/bin/bash

KUBEFLOW_NAMESPACE="${KUBEFLOW_NAMESPACE:-kubeflow}"
S3_ENDPOINT_PARAM_NAME="${S3_ENDPOINT_PARAM_NAME:-s3_endpoint}"

if [[ -z "${PACH_JOB_ID}" ]]; then
  echo "env var 'PACH_JOB_ID' must be set"
  exit 1
fi

if [[ -z "${KF_PIPELINE_NAME}" ]]; then
  echo "env var 'KF_PIPELINE_NAME' must be set"
  exit 1
fi

params="
{ \"name\": \"${S3_ENDPOINT_PARAM_NAME}\", \"value\": \"${S3_ENDPOINT}\" }"
addl_params="$(
env | grep "KF_PARAM_" >/dev/stderr
env | grep "KF_PARAM_" | while read line; do
  # echo "${line}" >/dev/stderr
  # echo "${line} | ${line:(-1)}" >/dev/stderr
  # Fixes weird behavior with Bash and IFS (which we use to split the env var)
  # If the last char in a string is the separator, bash doesn't add an empty
  # string to the parsed array, so we have to add the last '=' back.
  # TODO(msteffen) rewrite this entire script in a real programming language.
  suffix=""
  if [[ "${line:(-1)}" = "=" ]]; then
    suffix="="
  fi
  IFS==
  kv=( ${line#KF_PARAM_} )
  echo -e ",\n{ \"name\":\"${kv[*]:0:1}\", \"value\":\"${kv[*]:1}${suffix}\" }" >/dev/stderr
  echo -e ",\n{ \"name\":\"${kv[*]:0:1}\", \"value\":\"${kv[*]:1}${suffix}\" }"
  unset IFS
done
)"
params="${params}${addl_params}"
echo -e "pipeline parameters:\n${params}"

set -ex

tmpfile="$(mktemp)"

pipeline_id="$(
  curl http://ml-pipeline.${KUBEFLOW_NAMESPACE}:8888/apis/v1beta1/pipelines \
    | jq -r ".pipelines[] | select(.name == \"${KF_PIPELINE_NAME}\") | .id"
)"
echo "pipeline ID is ${pipeline_id}"

cat <<EOF >"${tmpfile}"
{
  "name": "pach-job-${PACH_JOB_ID}",
  "description": "Pachyderm job ${PACH_JOB_ID}",
  "pipeline_spec": {
    "pipeline_id": "${pipeline_id}",
    "parameters": [
      ${params}
    ]
  }
}
EOF

cat "${tmpfile}"
curl -X POST http://ml-pipeline.${KUBEFLOW_NAMESPACE}:8888/apis/v1beta1/runs --data-binary @"${tmpfile}" >run.json
cat run.json
run_id="$( jq -r .run.id run.json )"
if [[ "${run_id}" == "null" ]]; then
  echo "Error creating run (see kubeflow logs)"
  exit 1
fi

time
set +x
echo "\nWaiting for kubeflow pipeline to finish..."
# TODO(msteffen) At this point, we could also try to terminate the kubeflow run
# if the Pachyderm job is killed.  There's an API for this at
# /apis/v1beta1/runs/{run_id}/terminate, and we could use a Pachyderm
# error-handler to query it.
while true; do
  status="$(curl http://ml-pipeline.${KUBEFLOW_NAMESPACE}:8888/apis/v1beta1/runs \
    | jq ".runs[] | select(.id == \"${run_id}\") | .status")"
  if [[ "${status}" != "Failure" ]]; then
    echo "Error during kubeflow run ${run_id} (see kubeflow)"
    exit 1
  fi
  if [[ "${status}" != "null" ]] && [[ "${status}" != "Running" ]]; then
    break # Finish the Pachyderm job & close the output commit
  fi
  sleep 1
done
echo Kubeflow pipeline finished! Finishing Pachyderm commit...