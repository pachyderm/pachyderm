#!/bin/bash

set -euxo pipefail

# Make sure Pachyderm enterprise and auth are enabled
which aws || pip install awscli --upgrade --user

function activate {
    pachctl config update context $(pachctl config get active-context) --pachd-address=$(minikube ip):30650

    if [[ "$(pachctl enterprise get-state)" = "No Pachyderm Enterprise token was found" ]]; then
        # Don't print token to stdout
        # This is very important, or we'd leak it in our CI logs
        set +x
        pachctl enterprise activate "$ENT_ACT_CODE"
        set -x
    fi

    # Activate Pachyderm auth, if needed, and log in
    if ! pachctl auth list-admins; then
        admin="admin"
        echo "${admin}" | pachctl auth activate
    elif pachctl auth list-admins | grep "github:"; then
        admin="$( pachctl auth list-admins | grep 'github:' | head -n 1)"
        admin="${admin#github:}"
        echo "${admin}" | pachctl auth login
    else
        echo "Could not find a github user to log in as. Cannot get admin token"
        exit 1
    fi
}

function delete_all {
    if pachctl auth list-admins; then
        admin="$( pachctl auth list-admins | grep 'github:' | head -n 1)"
        admin="${admin#github:}"
        echo "${admin}" | pachctl auth login
    else
        echo "Could not find a github user to log in as. Cannot get admin token"
        exit 1
    fi
    echo "yes" | pachctl delete-all
}

eval "set -- $( getopt -l "activate,delete-all" "--" "${0}" "${@}" )"
while true; do
    case "${1}" in
     --activate)
        activate
        shift
        ;;
     --delete-all)
        delete_all
        shift
        ;;
     --)
        shift
        break
        ;;
     *)
        echo "Unrecognized operation: ${1}"
        echo
        echo "Operation should be \"--activate\" or \"--delete-all\""
        shift
        ;;
    esac
done



