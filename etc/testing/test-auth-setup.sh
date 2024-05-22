#!/bin/bash
set -euxo pipefail

# You can use this to set up 3 test users with different permissions on different projects
# Ellie: projectViewer fo joel-project
# Joel: projectWriter for ellie-project, projectViewer for bill-project
# Bill: projectViewer for ellie-project and joel-project
# All are projectOwners(can read/write) for their respective projects

export ROOT_TOKEN="db643083a0424318a2576049d3d60a7e" # not a real token, for local testing only
pachctl auth rotate-root-token --supply-token "$ROOT_TOKEN"
echo "$ROOT_TOKEN" | pachctl auth use-auth-token

# writer on joel project only
ELLIE_TOKEN=$(pachctl auth get-robot-token ellie --quiet)
# writer/viewer - can edit and view ellie-project, can only read bill project
JOEL_TOKEN=$(pachctl auth get-robot-token joel --quiet)
# joel viewer - can view joel project and repos but not edit 
BILL_TOKEN=$(pachctl auth get-robot-token bill --quiet)

# make projects
echo "$ELLIE_TOKEN" | pachctl auth use-auth-token
pachctl create project ellie-project
pachctl config update context --project ellie-project
(cd examples/opencv && make opencv)

echo "$JOEL_TOKEN" | pachctl auth use-auth-token
pachctl create project joel-project
pachctl config update context --project joel-project
(cd examples/opencv && make opencv)

echo "$BILL_TOKEN" | pachctl auth use-auth-token
pachctl create project bill-project # part of a group w/ reader
pachctl config update context --project bill-project
(cd examples/opencv && make opencv)

# set permissions on users
echo "$ROOT_TOKEN" | pachctl auth use-auth-token
pachctl auth set project joel-project projectViewer robot:ellie

pachctl auth set project ellie-project projectWriter robot:joel
pachctl auth set project bill-project projectViewer robot:joel

pachctl auth set project ellie-project projectViewer robot:bill
pachctl auth set project joel-project projectViewer robot:bill

echo "root: echo \"$ROOT_TOKEN\" | pachctl auth use-auth-token"
echo "ellie: echo \"$ELLIE_TOKEN\" | pachctl auth use-auth-token"
echo "joel: echo \"$JOEL_TOKEN\" | pachctl auth use-auth-token"
echo "bill: echo \"$BILL_TOKEN\" | pachctl auth use-auth-token"
