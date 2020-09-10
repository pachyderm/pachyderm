#!/bin/bash

NUM_PROCESSES=10

ID=$(ls /pfs/sample_data/ | head -n 1)
DATA_FOLDER="/pfs/sample_data/${ID}/"
INITIAL_EXAM_LIST_PATH="/pfs/sample_data/${ID}/gen_exam_list_before_cropping.pkl"

CROPPED_IMAGE_PATH="/pfs/out/${ID}/cropped_images"
CROPPED_EXAM_LIST_PATH="/pfs/out/${ID}/cropped_images/cropped_exam_list.pkl"
EXAM_LIST_PATH="/pfs/out/${ID}/data.pkl"
export PYTHONPATH=$(pwd):$PYTHONPATH

echo 'Stage 1: Crop Mammograms'
python3 src/cropping/crop_mammogram.py \
    --input-data-folder $DATA_FOLDER \
    --output-data-folder $CROPPED_IMAGE_PATH \
    --exam-list-path $INITIAL_EXAM_LIST_PATH  \
    --cropped-exam-list-path $CROPPED_EXAM_LIST_PATH  \
    --num-processes $NUM_PROCESSES