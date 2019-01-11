#!/bin/bash
# sort.sh creates a collection of pachyderm pipelines that look like:
#    left─┐
#         ├─copy─sort
#    right┘
#
# inputs:
# each input file is named for a digit and contains 2-digit numbers ending in
# that digit. E.g. '0' contains '00\n10\n20...', '1' contains '01\n11\n21\n...'
# left:  0,...,4 -> copy -> sort -> 01, 02, 03, 04, 05, ...
# right: 5,...,9

HERE="$(dirname "${0}")"
source "${HERE}/deploy.sh"

set -x

pachctl_1_7 create-repo left
pachctl_1_7 create-repo right

pachctl_1_7 create-pipeline -f - <<EOF
{
  "pipeline": {
    "name": "copy"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "cp /pfs/*/* /pfs/out"
    ]
  },
  "parallelism_spec": {
    "constant": 1
  },
  "input": {
    "union": [
      { "atom": { "repo": "left",  "glob": "/*" } },
      { "atom": { "repo": "right", "glob": "/*" } }
    ]
  },
  "enable_stats": true
}
{
  "pipeline": {
    "name": "sort"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "sort -n /pfs/copy/* >/pfs/out/nums"
    ]
  },
  "parallelism_spec": {
    "constant": 1
  },
  "input": { "atom": { "repo": "copy", "glob": "/" } },
  "enable_stats": true
}
EOF

tmpfile=$(mktemp -p.)
for _i in $(seq 0 9); do
  # roughly alternate between 'left' and 'right' by committing in a funny order
  # (this writes the files left/0, right/7, left/4, left/1, right/8, etc...)
  i=$(( _i*7 % 10 ))
  [[ "${i}" -ge 5 ]] && repo=right || repo=left

  # Clear tmpfile, write all two-digit numbers with ones place=$i to tmpfile
  echo -n "" >"${tmpfile}"
  for j in $(seq 0 9); do
    echo "${j}${i}" >>"${tmpfile}"
  done

  # Write to pachd
  pachctl_1_7 put-file "${repo}" master "/${i}" <"${tmpfile}"
done

# Wait for pipelines to process all commits
pachctl_1_7 flush-commit left/master

# Delete a few commits, as that has caused migration bugs in the past
# TODO(msteffen): Split this test up into tests of distinct bugs (stats,
# delete-commit, multiple pipelines)
pachctl_1_7 delete-commit left master~4
pachctl_1_7 delete-commit left master~3
pachctl_1_7 delete-commit right master~4
pachctl_1_7 delete-commit right master~3

echo "Extracting metadata from Pachyderm. Note that this step occasionally"
echo "fails due to transient encoding issues, and you may need to re-run it"

set -x

pachctl_1_7 extract >${HERE}/sort.metadata
