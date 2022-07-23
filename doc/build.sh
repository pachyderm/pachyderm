#!/bin/bash

set -ex

# Start out in same directory as script, so that relative paths below are all
# correct
cd "$(dirname "${0}")"

# Delete old site/ dir
if [[ -d site ]]; then
  rm -rf site overrides/partials/versions.html
fi

# Add each version of the docs to the dropdown defined by
# material/overrides/partials/versions.html. This must be built before running 'mkdocs'
# itself
latest_version="$(ls ./docs | grep -Ev 'latest|master|archived|2.3.x' | sort -r -V | head -n 1)"
cat <<EOF >overrides/partials/versions.html
<div class="mdl-selectfield">
    <select class="mdl-selectfield__select" id="version-selector" onchange="
        let pathParts = window.location.pathname.split('/');
        pathParts[1] = this.value;
        window.location.pathname = pathParts.join('/')
    ">
        <option style="color:white;background-color:#4b2a5c;" value="latest">latest (${latest_version})</option>
EOF

all_versions="$(ls ./docs | grep -Ev 'latest|master|archived|2.3.x' | sort -r -V)"
for d in $all_versions; do
    # don't link latest version again
    if [[ "${d}" == "$latest_version" ]]; then
         continue
    fi
    cat <<EOF >>overrides/partials/versions.html
        <option style="color:white;background-color:#4b2a5c;" value="${d}">${d}</option>"
EOF
done
   cat <<EOF >>overrides/partials/versions.html
        <option style="color:white;background-color:#4b2a5c;" value="2.3.x">2.3.x-alpha</option>"
EOF

    cat <<EOF >>overrides/partials/versions.html
        <option style="color:white;background-color:#4b2a5c;" value="archive">Archive</option>"
    </select>
    <!-- set initial value of 'select' to the version of the docs being browsed -->
    <script type="text/javascript">
      var pathParts = window.location.pathname.split('/');
      document.getElementById('version-selector').value = pathParts[1]
    </script>
</div>
EOF

# Rebuild all docs versions
#rm -rf site/*
for d in $all_versions; do
    out_dir="site/${d}"

    # Check for mkdocs file
    mkdocs_file="mkdocs-${d}.yml"
    if [[ ! -f "${mkdocs_file}" ]]; then
        echo "expected mkdocs config file \"${mkdocs_file}\" for docs version \"${d}\", but the config file wasn't found"
        exit 1
    fi

    # rebuild site
    mkdocs build --config-file "${mkdocs_file}" --site-dir "${out_dir}"
done
## Temp preview of 2.3.x-alpha out of master
#rm -rf ./site/2.3.x-alpha/
mkdocs build --config-file "mkdocs-2.3.x.yml" --site-dir "site/2.3.x"
# Finally, copy latest version of the docs into 'latest'
if [[ -z "${latest_version}" ]]; then
    echo "No latest version to symlink"
    exit 1
fi
cp -Rl "site/${latest_version}" site/latest

# Add custom 404
ln "site/${latest_version}/404.html" site/404.html
cp -Rl "site/${latest_version}/assets" site/assets
cp robots.txt site/robots.txt
cp google6522bd76c3ff2fe7.html site/google6522bd76c3ff2fe7.html