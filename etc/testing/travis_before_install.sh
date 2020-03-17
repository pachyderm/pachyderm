#!/bin/bash
if [[ -d $HOME/docker ]]; then
    ls $HOME/docker/*.tar.gz | xargs -I {file} sh -c "zcat {file} | docker load"
fi
