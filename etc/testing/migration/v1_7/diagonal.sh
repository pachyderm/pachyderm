#!/bin/bash
# diagonal.sh creates a contrived collection of pachyderm pipelines that look
# like:
#    left─┬─filter-left─┐
#         │             ├─join
#    right┴─filter-right┘
#
# inputs:
# left:  0,1,2,... -> 0/00,0/01,0/02,... -> 01, 02, 03, 04, 05, ...
# right: 0,1,2,...    0/00,0/10,0/20,...
#
# Due to https://github.com/pachyderm/pachyderm/issues/3337, restoring this
# cluster doesn't yet work (the cross input causes the restored pipeline to
# process commits with no matching datums)

HERE="$(dirname "${0}")"
# shellcheck source=./etc/testing/migration/v1_7/deploy.sh
source "${HERE}/deploy.sh"

set -x

pachctl_1_7 create-repo left
pachctl_1_7 create-repo right

pachctl_1_7 create-pipeline -f - <<EOF
{
  "pipeline": {
    "name": "filter-left"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "left=\$(ls /pfs/left)",
      "right=\$(ls /pfs/right)",
      "mkdir /pfs/out/\${left}",
      "touch /pfs/out/\${left}/\${left}\${right}"
    ]
  },
  "parallelism_spec": {
    "constant": 1
  },
  "input": {
    "cross": [
      {
        "atom": {
          "repo": "left",
          "glob": "/*"
        }
      },
      {
        "atom": {
          "repo": "right",
          "glob": "/*"
        }
      }
    ]
  },
  "enable_stats": true
}
{
  "pipeline": {
    "name": "filter-right"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "left=\$(ls /pfs/left)",
      "right=\$(ls /pfs/right)",
      "mkdir /pfs/out/\${right}",
      "touch /pfs/out/\${right}/\${left}\${right}"
    ]
  },
  "parallelism_spec": {
    "constant": 1
  },
  "input": {
    "cross": [
      {
        "atom": {
          "repo": "left",
          "glob": "/*"
        }
      },
      {
        "atom": {
          "repo": "right",
          "glob": "/*"
        }
      }
    ]
  },
  "enable_stats": true
}
{
  "pipeline": {
    "name": "join"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "find -L /pfs/left /pfs/right -type f -printf '%f\\\\n' \\\\",
      "| sort \\\\",
      "| uniq --repeated \\\\",
      "| xargs -I{} touch /pfs/out/{}"
    ]
  },
  "parallelism_spec": {
    "constant": 1
  },
  "input": {
    "cross": [
      {
        "atom": {
          "name": "left",
          "repo": "filter-left",
          "glob": "/*"
        }
      },
      {
        "atom": {
          "name": "right",
          "repo": "filter-right",
          "glob": "/*"
        }
      }
    ]
  },
  "enable_stats": true
}
EOF

for repo in left right; do
  for c in $(seq 0 9); do
    pachctl_1_7 put-file "${repo}" master "/${c}" </dev/null
  done
done

pachctl_1_7 flush-commit left/master

# Delete a few commits, as that has caused migration bugs in the past
# TODO(msteffen): Split this test up into tests of distinct bugs (stats,
# delete-commit, multiple pipelines)
pachctl_1_7 delete-commit left master~9
pachctl_1_7 delete-commit left master~8
pachctl_1_7 delete-commit left master~7
pachctl_1_7 delete-commit right master~9
pachctl_1_7 delete-commit right master~8
pachctl_1_7 delete-commit right master~7

pachctl_1_7 extract >"${HERE}/diagonal.dump"
