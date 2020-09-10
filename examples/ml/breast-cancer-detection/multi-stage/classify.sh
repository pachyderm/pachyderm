#!/bin/bash

DEVICE_TYPE='gpu'
NUM_EPOCHS=10
HEATMAP_BATCH_SIZE=100
GPU_NUMBER=0

ID=$(ls /pfs/crop/ | head -n 1)
IMAGEHEATMAPS_MODEL_PATH="/pfs/models/sample_imageheatmaps_model.p"

CROPPED_IMAGE_PATH="/pfs/crop/${ID}/cropped_images"
EXAM_LIST_PATH="/pfs/extract_centers/${ID}/data.pkl"
HEATMAPS_PATH="/pfs/generate_heatmaps/${ID}/heatmaps"
IMAGEHEATMAPS_PREDICTIONS_PATH="/pfs/out/${ID}/imageheatmaps_predictions.csv"
export PYTHONPATH=$(pwd):$PYTHONPATH

echo 'Stage 4b: Run Classifier (Image+Heatmaps)'
python3 src/modeling/run_model.py \
    --model-path $IMAGEHEATMAPS_MODEL_PATH \
    --data-path $EXAM_LIST_PATH \
    --image-path $CROPPED_IMAGE_PATH \
    --output-path $IMAGEHEATMAPS_PREDICTIONS_PATH \
    --use-heatmaps \
    --heatmaps-path $HEATMAPS_PATH \
    --use-augmentation \
    --num-epochs $NUM_EPOCHS \
    --device-type $DEVICE_TYPE \
    --gpu-number $GPU_NUMBER
