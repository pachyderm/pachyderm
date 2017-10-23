#!/bin/bash

set -x

pachctl delete-repo shakespeare-gcs
pachctl delete-repo shakespeare-local

rm -r ./shakespeare_test
mkdir -p shakespeare_test/{gcs,local}/{output,comparison}

pachctl create-repo shakespeare-gcs
COMMIT=$(pachctl start-commit shakespeare-gcs master | tee /dev/stderr)
# pachctl put-file shakespeare-gcs master -r -f gs://indx/Sequencing/Exome_seq/input
pachctl put-file shakespeare-gcs "${COMMIT}" / -r -f gs://indx/Sequencing/Exome_seq/input
pachctl finish-commit shakespeare-gcs "${COMMIT}"

pachctl create-repo shakespeare-local
COMMIT=$(pachctl start-commit shakespeare-local master | tee /dev/stderr)
for f in MB_1294_N_P0061.R1 MB_1294_N_P0061.R2 MB_1295_C_P0061.R1 MB_1295_C_P0061.R2; do
  gsutil cp "gs://indx/Sequencing/Exome_seq/input/${f}.fastq.gz" - | pv | pachctl put-file shakespeare-local ${COMMIT} /${f}.fastq.gz -f -
done
pachctl finish-commit shakespeare-local "${COMMIT}"

