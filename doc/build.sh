#!/bin/bash

mkdocs build --config-file mkdocs.yml --site-dir site/master/
mkdocs build --config-file mkdocs-1.9.x.yml --site-dir site/1.9.x/

cat <<EOF >material/partials/versions.html
<div class="mdl-selectfield">
<select class="mdl-selectfield__select" onchange="let pathParts = window.location.pathname.split('/ '); pathParts[1] = this.value; window.location.pathname = pathParts.join('/')">
EOF
for d in $(ls docs); do
    if [[ "${d}" == "archive" ]]; then
        continue
    fi
cat <<EOF >>material/partials/versions.html
              <option style="" selected value="${d}">${d}</option>
EOF
done
cat <<EOF >>material/partials/versions.html
     </select>
   </div>
EOF
