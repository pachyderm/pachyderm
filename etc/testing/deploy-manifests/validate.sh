#!/bin/bash

set -e

echo "Testing current pachctl against golden deployment manifests"
echo "Run"
echo "  validate.sh --regenerate"
echo "to replace golden deployment manifests with current output"
echo "(necessary if you have deliberately changed 'pachctl deploy')"
echo ""

here="$(dirname "${0}")"
dest_dir="test"
rm -rf "${here}/${dest_dir}" || true
mkdir -p "${here}/${dest_dir}"

is_regenerate=""
if [[ "${#@}" -eq 1 ]]; then
  if [[ "${1}" == "--regenerate" ]]; then
	  is_regenerate="true"
    dest_dir="golden"
  else
    echo "Unrecognized flag ${1}" >/dev/stderr
    echo "Must be --regenerate" >/dev/stderr
    exit 1
  fi
fi

# A custom deployment
custom_args=(
--secure
--dynamic-etcd-nodes 3
--etcd-storage-class storage-class
--namespace pachyderm
--no-expose-docker-socket
--object-store=s3
  pach-volume       # <volumes>
  50                # <size of volumes (in GB)>
  pach-bucket       # <bucket>
  storage-id        # <id>
  storage-secret    # <secret>
  storage.endpoint  # <endpoint>
)
google_args=(
--dynamic-etcd-nodes 3
  pach-bucket # <bucket-name>
  50          # <disk-size>
)
amazon_args=(
--dynamic-etcd-nodes 3
--credentials "AWSIDAWSIDAWSIDAWSID,awssecret+awssecret+awssecret+awssecret+"
  pach-bucket # <bucket-name>
  us-west-1   # <region>
  50          # <disk-size>
)
microsoft_args=(
--dynamic-etcd-nodes 3
  pach-container           # <container>
  pach-account             # <account-name>
  cGFjaC1hY2NvdW50LWtleQ== # <account-key> (base64-encoded "pach-account-key")
  50                       # <disk-size>
)

for plat in custom google amazon microsoft; do
  for fmt in json yaml; do
    output="${here}/${dest_dir}/${plat}-deploy-manifest.${fmt}"
    eval "args=( \"\${${plat}_args[@]}\" )"
    # Generate kubernetes manifest, and strip additional version info so that
    # pachctl builds from the same version all work
    pachctl deploy "${plat}" "${args[@]}" -o "${fmt}" --dry-run \
      | sed 's/\([0-9]\{1,4\}\.[0-9]\{1,4\}\.[0-9]\{1,4\}\)-[0-9a-f]\{40\}/\1/g' >"${output}"
    if [[ ! "${is_regenerate}" ]]; then
      # Check manifests with kubeval
      kubeval "${output}"
    fi
  done
done

# Compare manifests to golden files (in addition to kubeval, to see changes
# in storage secrets and such)
#
# TODO(msteffen): if we ever consider removing this because it generates too
# many spurious test failures, then I highly recomment we keep the 'kubeval'
# validation above, as it should accept any valid kubernetes manifest, and
# would've caught at least one serialization bug that completely broke 'pachctl
# deploy' in v1.9.8
if [[ ! "${is_regenerate}" ]]; then
  DIFF_CMD="${DIFF_CMD:-diff}"
  if ! "${DIFF_CMD}" "${here}/test" "${here}/golden"; then
    echo "Deployment manifest has changed." >/dev/stderr
    echo "If this deliberate, run:" >/dev/stderr
    echo "  validate.sh --regenerate" >/dev/stderr
    echo "  git add etc/testing/deploy-manifests/golden" >/dev/stderr
    echo "  git commit" >/dev/stderr
    echo "to replace golden deployment manifests with current output" >/dev/stderr
    exit 1
  fi
fi
