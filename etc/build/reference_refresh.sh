#!/bin/bash
# removes and regenerates docs/master/reference/pachctl on the current version of the doc
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${SCRIPT_DIR}/../govars.sh"

version="$("${PACHCTL}" version --client-only)"
major_minor=$(echo "$version" | cut -f -2 -d ".")
echo "--- Updating docs for version: $version"

# Set sed options for GNU/BSD sed
sed_opts=( "-i" )
if [[ "$(uname)" == "Darwin" ]]; then
  sed_opts+=( "" )
fi

# Rebuild pachctl docs
here="$(dirname "${0}")"
doc_root="${here}/../../doc"
pachctl_docs="${doc_root}/docs/${major_minor}.x/reference/pachctl"
rm -rf "${pachctl_docs}" && mkdir -p "${pachctl_docs}"
# TODO(msteffen): base this on $PACHCTL?
"${GOBIN}/pachctl-doc"  "${pachctl_docs}"

# Remove "see also" sections, since they have bad links and aren't very
# helpful here anyways
# LANG=C allows sed to ignore non-ascii files in the 'doc' directory on Mac (e.g. images)
LANG=C find "${pachctl_docs}" -name '*.md' -type f -exec \
  sed "${sed_opts[@]}" -n -e '/### SEE ALSO/,$d;p' {} \;


version_size="$(
  wc -c "${pachctl_docs}/pachctl_version.md" \
    | awk '{print $1}'
)"
if [[ "${version_size}" -lt 10 ]]; then
  echo "Error: auto-generated pachctl docs were somehow truncated or deleted"
  exit 1
fi
echo "--- Successfully updated docs"
