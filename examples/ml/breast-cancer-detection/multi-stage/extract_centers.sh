#!/bin/bash

NUM_PROCESSES=10

ID=$(ls /pfs/crop/ | head -n 1)

CROPPED_IMAGE_PATH="/pfs/crop/${ID}/cropped_images"
CROPPED_EXAM_LIST_PATH="/pfs/crop/${ID}/cropped_images/cropped_exam_list.pkl"
EXAM_LIST_PATH="/pfs/out/${ID}/data.pkl"
export PYTHONPATH=$(pwd):$PYTHONPATH

echo 'Stage 2: Extract Centers'
python3 src/optimal_centers/get_optimal_centers.py \
    --cropped-exam-list-path $CROPPED_EXAM_LIST_PATH \
    --data-prefix $CROPPED_IMAGE_PATH \
    --output-exam-list-path $EXAM_LIST_PATH \
    --num-processes $NUM_PROCESSES