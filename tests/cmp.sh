#!/bin/bash

if [[ "$#" -ne 2 ]]; then
  echo "Must provide two files to compare"
  exit 1
fi

set -x

tmp=$( mktemp -d $(pwd)/split_txt.XXXXXXXXXX )

split -d -n 1000  "${1}" ${tmp}/a
size="$( du -b ${tmp}/a000 | awk '{print $1}' )"
split -d -a 3 -b "${size}" "${2}" ${tmp}/b

set +x

for i in $(seq -w 0 999); do
  md5sum ${tmp}/a${i} | awk '{print $1}' >> ${tmp}/a_hashes.txt
  md5sum ${tmp}/b${i} | awk '{print $1}' >> ${tmp}/b_hashes.txt
done

diff -y ${tmp}/a_hashes.txt ${tmp}/b_hashes.txt | awk '{ printf("%03d.  %s\n", NR-1, $0) }' >${tmp}/diff.txt
