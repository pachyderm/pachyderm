#!/bin/bash
# Delete site/
rm -rf site/
# Build each version of the docs
mkdocs build --config-file mkdocs.yml --site-dir site/latest/
mkdocs build --config-file mkdocs-1.9.x.yml --site-dir site/1.9.x/

# Add each version of the docs to the dropdown defined by
# "material/partials/versions.html"
cat <<EOF >material/partials/versions.html
<div class="mdl-selectfield">
        <select class="mdl-selectfield__select" onchange="let pathParts = window.location.pathname.split('/'); pathParts[1] = this. value; window.location.pathname = pathParts.join('/')">
               <option>Select a version</option>
               <option style="" value="latest">latest</option>

EOF
for d in $(ls docs); do
    if [[ "${d}" == "archive" ]]; then
        continue
    fi
    if [[ "${d}" == "latest" ]]; then
         continue
    fi
cat <<EOF >>material/partials/versions.html
              <option style="" value="${d}">${d}</option>

EOF
done
cat <<EOF >>material/partials/versions.html
     </select>
   </div>
EOF
