#!/bin/bash

set -e

version="$("$GOPATH/bin/pachctl" version --client-only)"
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
pachctl_docs="${doc_root}/docs/master/reference/pachctl"
rm -rf "${pachctl_docs}" && mkdir "${pachctl_docs}"
"${GOPATH}/bin/pachctl-doc"

# Remove "see also" sections, since they have bad links and aren't very
# helpful here anyways
# LANG=C allows sed to ignore non-ascii files in the 'doc' directory on Mac (e.g. images)
LANG=C find "${pachctl_docs}" -name '*.md' -type f -exec \
  sed "${sed_opts[@]}" -n -e '/### SEE ALSO/,$d;p' {} \;

# Copy "master" to current version's docs
rm -rf "${doc_root}/docs/${major_minor}.x"
cp -R "${doc_root}/docs/master" "${doc_root}/docs/${major_minor}.x"

# Copy navigation from "master" mkdocs.yml to current version mkdocs.yml
cp "${doc_root}/mkdocs-master.yml" "${doc_root}/mkdocs-${major_minor}.x.yml"
sed "${sed_opts[@]}" -e "s#docs_dir: docs/master/#docs_dir: docs/${major_minor}.x#g" "${doc_root}/mkdocs-${major_minor}.x.yml"
sed "${sed_opts[@]}" -e "s#site_dir: site/master/#site_dir: site/${major_minor}.x#g" "${doc_root}/mkdocs-${major_minor}.x.yml"

version_size="$(
  wc -c "${doc_root}/docs/master/reference/pachctl/pachctl_version.md" \
    | awk '{print $1}'
)"
if [[ "${version_size}" -lt 10 ]]; then
  echo "Error: auto-generated pachctl docs were somehow truncated or deleted"
  exit 1
fi
echo "--- Successfully updated docs"
