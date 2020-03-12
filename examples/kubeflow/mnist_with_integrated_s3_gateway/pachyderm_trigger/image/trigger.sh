#!/bin/bash

KUBEFLOW_NAMESPACE="${KUBEFLOW_NAMESPACE:-kubeflow}"
S3_ENDPOINT_PARAM_NAME="${S3_ENDPOINT_PARAM_NAME:-s3_endpoint}"

if [[ -z "${PACH_JOB_ID}" ]]; then
  echo "env var 'PACH_JOB_ID' must be set"
  exit 1
fi

if [[ -z "${PIPELINE_NAME}" ]]; then
  echo "env var 'PIPELINE_NAME' must be set"
  exit 1
fi

params="$(
env | grep "KF_PARAM_" | while read line; do
  echo "${line} | ${line:(-1)}" >/dev/stderr
  suffix=""
  if [[ "${line:(-1)}" = "=" ]]; then
    suffix="="
  fi
  IFS==
  kv=( ${line#KF_PARAM_} )
  echo "{\"name\":\"${kv[*]:0:1}\",\"value\":\"${kv[*]:1}${suffix}\"}"
  unset IFS
done
)"
params="$( echo "${params}" | jq .)"
echo -e "pipeline parameters:\n${params}"

set -ex

tmpfile="$(mktemp)"

pipeline_id="$(
  curl http://ml-pipeline.${KUBEFLOW_NAMESPACE}:8888/apis/v1beta1/pipelines \
    | jq -r ".pipelines[] | select(.name == \"${PIPELINE_NAME}\") | .id"
)"
echo "pipeline ID is ${pipeline_id}"

cat <<EOF >"${tmpfile}"
{
  "name": "pach-job-${PACH_JOB_ID}",
  "description": "Pachyderm job ${PACH_JOB_ID}",
  "pipeline_spec": {
    "pipeline_id": "${pipeline_id}",
    "parameters": [
      {
        "name": "${S3_ENDPOINT_PARAM_NAME}",
        "value": "${S3_ENDPOINT}"
      }
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
echo Waiting for kubeflow pipeline to finish...
# TODO(msteffen) At this point, we could also try to terminate the kubeflow run
# if the Pachyderm job is killed.  There's an API for this at
# /apis/v1beta1/runs/{run_id}/terminate, and we could use a Pachyderm
# error-handler to query it.
while true; do
  status="$(curl POST http://ml-pipeline.${KUBEFLOW_NAMESPACE}:8888/apis/v1beta1/runs \
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
