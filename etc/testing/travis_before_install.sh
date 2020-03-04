#!/bin/bash
if [[ -d $HOME/docker ]]; then
    find "$HOME/docker" -name "*.tar.gz" -print0 | xargs -I "{file}" sh -c "zcat {file} | docker load"
fi
