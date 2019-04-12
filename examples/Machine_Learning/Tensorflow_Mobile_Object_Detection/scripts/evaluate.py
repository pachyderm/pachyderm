#!/usr/bin/python
#
# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import os

import sys
import argparse

import numpy as np
import PIL.Image as Image
import tensorflow as tf 

import scripts.retrain as retrain
from scripts.count_ops import load_graph

def evaluate_graph(graph_file_name):
    with load_graph(graph_file_name).as_default() as graph:
        ground_truth_input = tf.placeholder(
            tf.float32, [None, 5], name='GroundTruthInput')
        
        image_buffer_input = graph.get_tensor_by_name('input:0')
        final_tensor = graph.get_tensor_by_name('final_result:0')
        accuracy, _ = retrain.add_evaluation_step(final_tensor, ground_truth_input)
        
        logits = graph.get_tensor_by_name("final_training_ops/Wx_plus_b/add:0")
        xent = tf.reduce_mean(tf.nn.softmax_cross_entropy_with_logits(
            labels = ground_truth_input,
            logits = logits))
        
    image_dir = 'tf_files/flower_photos'
    testing_percentage = 10
    validation_percentage = 10
    validation_batch_size = 100
    category='testing'
    
    image_lists = retrain.create_image_lists(
        image_dir, testing_percentage,
        validation_percentage)
    class_count = len(image_lists.keys())
    
    ground_truths = []
    filenames = []
        
    for label_index, label_name in enumerate(image_lists.keys()):
      for image_index, image_name in enumerate(image_lists[label_name][category]):
        image_name = retrain.get_image_path(
            image_lists, label_name, image_index, image_dir, category)
        ground_truth = np.zeros([1, class_count], dtype=np.float32)
        ground_truth[0, label_index] = 1.0
        ground_truths.append(ground_truth)
        filenames.append(image_name)
    
    accuracies = []
    xents = []
    with tf.Session(graph=graph) as sess:
        for filename, ground_truth in zip(filenames, ground_truths):    
            image = Image.open(filename).resize((224,224),Image.ANTIALIAS)
            image = np.array(image, dtype=np.float32)[None,...]
            image = (image-128)/128.0

            feed_dict={
                image_buffer_input: image,
                ground_truth_input: ground_truth}

            eval_accuracy, eval_xent = sess.run([accuracy, xent], feed_dict)
            
            accuracies.append(eval_accuracy)
            xents.append(eval_xent)
        
        
    return np.mean(accuracies), np.mean(xents)

if __name__ == "__main__":
    os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'
    accuracy,xent = evaluate_graph(*sys.argv[1:])
    print('Accuracy: %g' % accuracy)
    print('Cross Entropy: %g' % xent)

    
  
