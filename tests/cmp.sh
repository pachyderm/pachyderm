#!/bin/bash

############
# Parse flags
############
COUNT=1000
eval "set -- $( getopt -l "count:" -- "${0}" "${@}" )"
while true; do
    case "${1}" in
        --count)
          COUNT="${2}"
          shift 2
          ;;
        --)
          shift
          break
          ;;
    esac
done

if [[ "$#" -ne 2 ]]; then
  echo "Must provide two files to compare"
  exit 1
fi

set -x

############
# make tmp directory, and split up files into aXXXXX and bXXXXXX in that
# directory.
############
tmp=$( mktemp -d $(pwd)/split_txt.XXXXXXXXXX )

split -d -a 5 -n ${COUNT}  "${1}" ${tmp}/a
size="$( du -b ${tmp}/a00000 | awk '{print $1}' )"
split -d -a 5 -b "${size}" "${2}" ${tmp}/b

set +x

############
# Hash aXXXXX into a_hashes.txt, and likewise for b_hashes.txt
############
for i in $(seq -f '%05.0f' $(( ${COUNT} - 1 )) ); do
  md5sum ${tmp}/a${i} | awk '{print $1}' >> ${tmp}/a_hashes.txt
  md5sum ${tmp}/b${i} | awk '{print $1}' >> ${tmp}/b_hashes.txt
done

############
# Create a side-by-side diff of the hashes
############
diff -y ${tmp}/a_hashes.txt ${tmp}/b_hashes.txt | awk '{ printf("%03d.\t%s\n", NR, $0) }' >${tmp}/diff.txt

