// Copyright 2017 The TensorFlow Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#import <AVFoundation/AVFoundation.h>
#import <UIKit/UIKit.h>
#import <Photos/Photos.h>

#include <vector>

#include "tensorflow/contrib/lite/kernels/register.h"
#include "tensorflow/contrib/lite/model.h"

typedef struct {
  int width;
  int height;
  int channels;
  std::vector<uint8_t> data;
} image_data;


@interface CameraExampleViewController
    : UIViewController<UIGestureRecognizerDelegate, AVCaptureVideoDataOutputSampleBufferDelegate> {
  IBOutlet UIView* previewView;
  CALayer* previewLayer;
  UIView* flashView;
  NSMutableArray* labelLayers;


  std::vector<std::string> labels;
  std::unique_ptr<tflite::FlatBufferModel> model;
  tflite::ops::builtin::BuiltinOpResolver resolver;
  std::unique_ptr<tflite::Interpreter> interpreter;

  double total_latency;
  int total_count;
        
  int photos_index;
  PHFetchResult *photos;
  
  UIImage* input_image;
  UIImage* display_image;
}
@property(strong, nonatomic) CATextLayer* predictionTextLayer;

- (IBAction)takePicture:(id)sender;
@end

