#!/bin/bash

cat <<EOF
Files:
  https://storage.googleapis.com/indx/Sequencing/Exome_seq/input/MB_1294_N_P0061.R1.fastq.gz
  https://storage.googleapis.com/indx/Sequencing/Exome_seq/input/MB_1294_N_P0061.R2.fastq.gz
  https://storage.googleapis.com/indx/Sequencing/Exome_seq/input/MB_1295_C_P0061.R1.fastq.gz
  https://storage.googleapis.com/indx/Sequencing/Exome_seq/input/MB_1295_C_P0061.R2.fastq.gz
EOF

set -x

pachctl create-repo gene-bad
COMMIT=$(pachctl start-commit gene-bad master | tee /dev/stderr)
# pachctl put-file gene-bad master -r -f gs://indx/Sequencing/Exome_seq/input
pachctl put-file gene-bad "${COMMIT}" / -r -f gs://indx/Sequencing/Exome_seq/input
pachctl finish-commit gene-bad "${COMMIT}"

pachctl create-repo gene-good
COMMIT=$(pachctl start-commit gene-good master | tee /dev/stderr)
for f in MB_1294_N_P0061.R1 MB_1294_N_P0061.R2 MB_1295_C_P0061.R1 MB_1295_C_P0061.R2; do
  gsutil cp "gs://indx/Sequencing/Exome_seq/input/${f}.fastq.gz" - | pv | pachctl put-file gene-good ${COMMIT} /${f}.fastq.gz -f -
done
pachctl finish-commit gene-good "${COMMIT}"

