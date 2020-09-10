#!/bin/bash

DEVICE_TYPE='gpu'
HEATMAP_BATCH_SIZE=100
GPU_NUMBER=0

ID=$(ls /pfs/crop/ | head -n 1)
PATCH_MODEL_PATH="/pfs/models/sample_patch_model.p"

CROPPED_IMAGE_PATH="/pfs/crop/${ID}/cropped_images"
EXAM_LIST_PATH="/pfs/extract_centers/${ID}/data.pkl"
HEATMAPS_PATH="/pfs/out/${ID}/heatmaps"
export PYTHONPATH=$(pwd):$PYTHONPATH

echo 'Stage 3: Generate Heatmaps'
python3 src/heatmaps/run_producer.py \
    --model-path $PATCH_MODEL_PATH \
    --data-path $EXAM_LIST_PATH \
    --image-path $CROPPED_IMAGE_PATH \
    --batch-size $HEATMAP_BATCH_SIZE \
    --output-heatmap-path $HEATMAPS_PATH \
    --device-type $DEVICE_TYPE \
    --gpu-number $GPU_NUMBER

