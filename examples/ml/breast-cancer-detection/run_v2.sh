#!/bin/bash

NUM_PROCESSES=10
DEVICE_TYPE=$1
NUM_EPOCHS=10
HEATMAP_BATCH_SIZE=100
GPU_NUMBER=0

DATA_FOLDER='/pfs/sample_data/'
INITIAL_EXAM_LIST_PATH='/pfs/sample_data/exam_list_before_cropping.pkl'
PATCH_MODEL_PATH='/pfs/models/sample_patch_model.p'
IMAGE_MODEL_PATH='/pfs/models/sample_image_model.p'
IMAGEHEATMAPS_MODEL_PATH='/pfs/models/sample_imageheatmaps_model.p'

CROPPED_IMAGE_PATH='/pfs/out/cropped_images'
CROPPED_EXAM_LIST_PATH='/pfs/out/cropped_images/cropped_exam_list.pkl'
EXAM_LIST_PATH='/pfs/out/data.pkl'
HEATMAPS_PATH='/pfs/out/heatmaps'
IMAGE_PREDICTIONS_PATH='/pfs/out/image_predictions.csv'
IMAGEHEATMAPS_PREDICTIONS_PATH='/pfs/out/imageheatmaps_predictions.csv'
export PYTHONPATH=$(pwd):$PYTHONPATH

echo 'Stage 1: Crop Mammograms'
python3 src/cropping/crop_mammogram.py \
    --input-data-folder $DATA_FOLDER \
    --output-data-folder $CROPPED_IMAGE_PATH \
    --exam-list-path $INITIAL_EXAM_LIST_PATH  \
    --cropped-exam-list-path $CROPPED_EXAM_LIST_PATH  \
    --num-processes $NUM_PROCESSES

echo 'Stage 2: Extract Centers'
python3 src/optimal_centers/get_optimal_centers.py \
    --cropped-exam-list-path $CROPPED_EXAM_LIST_PATH \
    --data-prefix $CROPPED_IMAGE_PATH \
    --output-exam-list-path $EXAM_LIST_PATH \
    --num-processes $NUM_PROCESSES

echo 'Stage 3: Generate Heatmaps'
python3 src/heatmaps/run_producer.py \
    --model-path $PATCH_MODEL_PATH \
    --data-path $EXAM_LIST_PATH \
    --image-path $CROPPED_IMAGE_PATH \
    --batch-size $HEATMAP_BATCH_SIZE \
    --output-heatmap-path $HEATMAPS_PATH \
    --device-type $DEVICE_TYPE \
    --gpu-number $GPU_NUMBER

echo 'Stage 4a: Run Classifier (Image)'
python3 src/modeling/run_model.py \
    --model-path $IMAGE_MODEL_PATH \
    --data-path $EXAM_LIST_PATH \
    --image-path $CROPPED_IMAGE_PATH \
    --output-path $IMAGE_PREDICTIONS_PATH \
    --use-augmentation \
    --num-epochs $NUM_EPOCHS \
    --device-type $DEVICE_TYPE \
    --gpu-number $GPU_NUMBER

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
