#!/bin/bash

set -ex

# Create 'latest' symlink
latest_version="1.9.x"
if [[ -L site/latest ]]; then
  rm site/latest
fi
ln -s "${latest_version}" site/latest

# Rebuild all docs versions
for d in $(ls docs); do
    # don't rebuild archive dir
    if [[ "${d}" == "archive" ]]; then
        continue
    fi
    in_dir="./docs/${d}"
    out_dir="./site/${d}"

    # Check for mkdocs file
    mkdocs_file="mkdocs-${d}.yml"
    if [[ ! -f "${mkdocs_file}" ]]; then
        echo "expected mkdocs config file \"${mkdocs_file}\" for docs version \"${d}\", but the config file wasn't found"
        exit 1
    fi

    # # Skip building docs if the version has changed
    # dir_hash="$( find "${in_dir}" -type f | xargs md5sum | md5sum )"
    # old_hash="-"
    # if [[ -f "${out_dir}/checksum" ]]; then
    #     old_hash="$( cat "${out_dir}/checksum" )"
    # fi
    # if [[ "${dir_hash}" == "${old_hash}" ]]; then
    #     continue
    # fi
    # # Record new hash
    # echo "${dir_hash}" >"${out_dir}/checksum"
    # Delete old built site/
    if [[ -d  "${out_dir}" ]]; then
      rm -r "${out_dir}"
    fi

    # rebuild site
    mkdocs build --config-file "${mkdocs_file}" --site-dir "${out_dir}"
done

# Add each version of the docs to the dropdown defined by
# "material/partials/versions.html"
cat <<EOF >material/partials/versions.html
<div class="mdl-selectfield">
<select class="mdl-selectfield__select" id="version-selector" onchange="let pathParts = window.location.pathname.split('/'); pathParts[1] = this.value; window.location.pathname = pathParts.join('/')">
EOF
for d in $(ls docs); do
    # don't rebuild archive dir
    if [[ "${d}" == "archive" ]]; then
        continue
    fi
    cat <<EOF >>material/partials/versions.html
        <option value="${d}">${d}</option>
EOF
done
cat <<EOF >>material/partials/versions.html
    </select>
    <!-- set initial value of 'select' to the version of the docs being browsed -->
    <script type="text/javascript">
      var pathParts = window.location.pathname.split('/');
      document.getElementById('version-selector').value = pathParts[1]
    </script>
   </div>
EOF
