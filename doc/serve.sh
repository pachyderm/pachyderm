#!/bin/bash

mkdocs serve --config-file mkdocs.yml
mkdocs serve --config-file mkdocs-1.9.x.yml

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
