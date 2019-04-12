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

#import "CameraExampleViewController.h"
#import <AssertMacros.h>
#import <AssetsLibrary/AssetsLibrary.h>
#import <CoreImage/CoreImage.h>
#import <ImageIO/ImageIO.h>

#include <sys/time.h>
#include <fstream>
#include <iostream>
#include <queue>

#include "tensorflow/contrib/lite/kernels/register.h"
#include "tensorflow/contrib/lite/model.h"
#include "tensorflow/contrib/lite/string_util.h"
#include "tensorflow/contrib/lite/mutable_op_resolver.h"

#define LOG(x) std::cerr

// If you have your own model, modify this to the file name, and make sure
// you've added the file to your app resources too.
static NSString* model_file_name = @"graph";
static NSString* model_file_type = @"lite";

// If you have your own model, point this to the labels file.
static NSString* labels_file_name = @"labels";
static NSString* labels_file_type = @"txt";

// These dimensions need to match those the model was trained with.
static const int wanted_input_width = 224;
static const int wanted_input_height = 224;
static const int wanted_input_channels = 3;

static NSString* FilePathForResourceName(NSString* name, NSString* extension) {
  NSString* file_path = [[NSBundle mainBundle] pathForResource:name ofType:extension];
  if (file_path == NULL) {
    LOG(FATAL) << "Couldn't find '" << [name UTF8String] << "." << [extension UTF8String]
               << "' in bundle.";
  }
  return file_path;
}

static void LoadLabels(NSString* file_name, NSString* file_type,
                       std::vector<std::string>* label_strings) {
  NSString* labels_path = FilePathForResourceName(file_name, file_type);
  if (!labels_path) {
    LOG(ERROR) << "Failed to find model proto at" << [file_name UTF8String]
               << [file_type UTF8String];
  }
  std::ifstream t;
  t.open([labels_path UTF8String]);
  std::string line;
  while (t) {
    std::getline(t, line);
    if (line.length()){
      label_strings->push_back(line);
    }
  }
  t.close();
}

// Returns the top N confidence values over threshold in the provided vector,
// sorted by confidence in descending order.
static void GetTopN(const float* prediction, const int prediction_size, const int num_results,
                    const float threshold, std::vector<std::pair<float, int>>* top_results) {
  // Will contain top N results in ascending order.
  std::priority_queue<std::pair<float, int>, std::vector<std::pair<float, int>>,
                      std::greater<std::pair<float, int>>>
      top_result_pq;

  const long count = prediction_size;
  for (int i = 0; i < count; ++i) {
    const float value = prediction[i];
    // Only add it if it beats the threshold and has a chance at being in
    // the top N.
    if (value < threshold) {
      continue;
    }

    top_result_pq.push(std::pair<float, int>(value, i));

    // If at capacity, kick the smallest value out.
    if (top_result_pq.size() > num_results) {
      top_result_pq.pop();
    }
  }

  // Copy to output vector and reverse into descending order.
  while (!top_result_pq.empty()) {
    top_results->push_back(top_result_pq.top());
    top_result_pq.pop();
  }
  std::reverse(top_results->begin(), top_results->end());
}

@interface CameraExampleViewController (InternalMethods)
- (void)teardownAVCapture;
@end

@implementation CameraExampleViewController

- (void) attachPreviewLayer{
  photos_index = 0;
  photos = nil;
  previewLayer = [[CALayer alloc] init];
  
  [previewLayer setBackgroundColor:[[UIColor blackColor] CGColor]];
  CALayer* rootLayer = [previewView layer];
  [rootLayer setMasksToBounds:YES];
  [previewLayer setFrame:[rootLayer bounds]];
  [rootLayer addSublayer:previewLayer];
  
  [self UpdatePhoto];
}

- (void)UpdatePhoto{
  PHAsset* asset;
  if (photos==nil || photos_index >= photos.count){
    [self updatePhotosLibrary];
    photos_index=0;
  }
  if (photos.count){
    asset = photos[photos_index];
    photos_index += 1;
    input_image = [self convertImageFromAsset:asset
                                   targetSize:CGSizeMake(wanted_input_width, wanted_input_height)
                                         mode:PHImageContentModeAspectFill];
    display_image = [self convertImageFromAsset:asset
                                     targetSize:CGSizeMake(asset.pixelWidth,asset.pixelHeight)
                                           mode:PHImageContentModeAspectFit];
    [self DrawImage];
  }
  
  if (input_image != nil){
    image_data image = [self CGImageToPixels:input_image.CGImage];
    [self inputImageToModel:image];
    [self runModel];
  }
}

- (void)DrawImage{
  CGFloat view_height = 800;
  CGFloat view_width = 600;
  
  UIGraphicsBeginImageContextWithOptions(CGSizeMake(view_width, view_height), NO, 0.0);
  CGContextRef context = UIGraphicsGetCurrentContext();
  UIGraphicsPushContext(context);
  
  float scale = view_width/display_image.size.width;
  
  if (display_image.size.height*scale > view_height){
    scale = view_height/display_image.size.height;
  }
  
  CGPoint origin = CGPointMake((view_width - display_image.size.width*scale) / 2.0f,
                               (view_height - display_image.size.height*scale) / 2.0f);
  [display_image drawInRect:CGRectMake(origin.x, origin.y,
                                       display_image.size.width*scale,
                                       display_image.size.height*scale)];
  UIGraphicsPopContext();
  display_image = UIGraphicsGetImageFromCurrentImageContext();
  UIGraphicsEndImageContext();
  previewLayer.contents = (id) display_image.CGImage;
}

- (void)teardownAVCapture {
  [previewLayer removeFromSuperlayer];
}

- (void) updatePhotosLibrary{
  PHFetchOptions *fetchOptions = [[PHFetchOptions alloc] init];
  fetchOptions.sortDescriptors = @[[NSSortDescriptor sortDescriptorWithKey:@"creationDate" ascending:YES]];
  photos = [PHAsset fetchAssetsWithMediaType:PHAssetMediaTypeImage options:fetchOptions];
}

- (UIImage *) convertImageFromAsset:(PHAsset *)asset
                         targetSize:(CGSize) targetSize
                              mode:(PHImageContentMode) mode{
  PHImageManager * manager = [[PHImageManager alloc] init];
  PHImageRequestOptions * options = [[PHImageRequestOptions alloc] init];
  NSMutableArray * images = [[NSMutableArray alloc] init];
  NSMutableArray * infos = [[NSMutableArray alloc] init];
  
  options.synchronous = TRUE;
      
  [manager requestImageForAsset:asset
                     targetSize:targetSize
                    contentMode:mode
                        options:options
                  resultHandler:^(UIImage *image, NSDictionary *info){
                                  [images addObject:image];
                                  [infos addObject:info];
                                }
  ];
  
  UIImage *result = images[0];

  return result;
}

- (AVCaptureVideoOrientation)avOrientationForDeviceOrientation:
    (UIDeviceOrientation)deviceOrientation {
  AVCaptureVideoOrientation result = (AVCaptureVideoOrientation)(deviceOrientation);
  if (deviceOrientation == UIDeviceOrientationLandscapeLeft)
    result = AVCaptureVideoOrientationLandscapeRight;
  else if (deviceOrientation == UIDeviceOrientationLandscapeRight)
    result = AVCaptureVideoOrientationLandscapeLeft;
  return result;
}

- (image_data)CGImageToPixels:(CGImage *)image {
  image_data result;
  result.width = (int)CGImageGetWidth(image);
  result.height = (int)CGImageGetHeight(image);
  result.channels = 4;
  
  CGColorSpaceRef color_space = CGColorSpaceCreateDeviceRGB();
  const int bytes_per_row = (result.width * result.channels);
  const int bytes_in_image = (bytes_per_row * result.height);
  result.data = std::vector<uint8_t>(bytes_in_image);
  const int bits_per_component = 8;
  
  CGContextRef context =
    CGBitmapContextCreate(result.data.data(), result.width, result.height, bits_per_component, bytes_per_row,
                          color_space, kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big);
  CGColorSpaceRelease(color_space);
  CGContextDrawImage(context, CGRectMake(0, 0, result.width, result.height), image);
  CGContextRelease(context);
  
  return result;
}



- (IBAction)takePicture:(id)sender {
    [self UpdatePhoto];
}

- (void)inputImageToModel:(image_data)image{
  float* out = interpreter->typed_input_tensor<float>(0);
  
  const float input_mean = 127.5f;
  const float input_std = 127.5f;
  assert(image.channels >= wanted_input_channels);
  uint8_t* in = image.data.data();

  for (int y = 0; y < wanted_input_height; ++y) {
    const int in_y = (y * image.height) / wanted_input_height;
    uint8_t* in_row = in + (in_y * image.width * image.channels);
    float* out_row = out + (y * wanted_input_width * wanted_input_channels);
    for (int x = 0; x < wanted_input_width; ++x) {
      const int in_x = (x * image.width) / wanted_input_width;
      uint8_t* in_pixel = in_row + (in_x * image.channels);
      float* out_pixel = out_row + (x * wanted_input_channels);
      for (int c = 0; c < wanted_input_channels; ++c) {
        out_pixel[c] = (in_pixel[c] - input_mean) / input_std;
      }
    }
  }
}

- (void)runModel {
  double startTimestamp = [[NSDate new] timeIntervalSince1970];
  if (interpreter->Invoke() != kTfLiteOk) {
    LOG(FATAL) << "Failed to invoke!";
  }
  double endTimestamp = [[NSDate new] timeIntervalSince1970];
  total_latency += (endTimestamp - startTimestamp);
  total_count += 1;
  NSLog(@"Time: %.4lf, avg: %.4lf, count: %d", endTimestamp - startTimestamp,
        total_latency / total_count,  total_count);

  const int output_size = (int)labels.size();
  const int kNumResults = 5;
  const float kThreshold = 0.1f;

  std::vector<std::pair<float, int>> top_results;

  float* output = interpreter->typed_output_tensor<float>(0);
  GetTopN(output, output_size, kNumResults, kThreshold, &top_results);
  
  std::vector<std::pair<float, std::string>> newValues;
  for (const auto& result : top_results) {
    std::pair<float, std::string> item;
    item.first = result.first;
    item.second = labels[result.second];
    
    newValues.push_back(item);
  }
  dispatch_async(dispatch_get_main_queue(), ^(void) {
    [self setPredictionValues:newValues];
  });
}

- (void)dealloc {
  [self teardownAVCapture];
}

- (void)didReceiveMemoryWarning {
  [super didReceiveMemoryWarning];
}

- (void)viewDidLoad {
  [super viewDidLoad];
  labelLayers = [[NSMutableArray alloc] init];

  NSString* graph_path = FilePathForResourceName(model_file_name, model_file_type);
  model = tflite::FlatBufferModel::BuildFromFile([graph_path UTF8String]);
  if (!model) {
    LOG(FATAL) << "Failed to mmap model " << graph_path;
  }
  LOG(INFO) << "Loaded model " << graph_path;
  model->error_reporter();
  LOG(INFO) << "resolved reporter";

  tflite::ops::builtin::BuiltinOpResolver resolver;
  LoadLabels(labels_file_name, labels_file_type, &labels);

  tflite::InterpreterBuilder(*model, resolver)(&interpreter);
  if (!interpreter) {
    LOG(FATAL) << "Failed to construct interpreter";
  }
  if (interpreter->AllocateTensors() != kTfLiteOk) {
    LOG(FATAL) << "Failed to allocate tensors!";
  }

  [self attachPreviewLayer];
}

- (void)viewDidUnload {
  [super viewDidUnload];
}

- (void)viewWillAppear:(BOOL)animated {
  [super viewWillAppear:animated];
}

- (void)viewDidAppear:(BOOL)animated {
  [super viewDidAppear:animated];
}

- (void)viewWillDisappear:(BOOL)animated {
  [super viewWillDisappear:animated];
}

- (void)viewDidDisappear:(BOOL)animated {
  [super viewDidDisappear:animated];
}

- (BOOL)shouldAutorotateToInterfaceOrientation:(UIInterfaceOrientation)interfaceOrientation {
  return (interfaceOrientation == UIInterfaceOrientationPortrait);
}

- (BOOL)prefersStatusBarHidden {
  return YES;
}

- (void)setPredictionValues:(std::vector<std::pair<float, std::string>>)newValues {

  const float leftMargin = 10.0f;
  const float topMargin = 10.0f;

  const float valueWidth = 48.0f;
  const float valueHeight = 18.0f;

  const float labelWidth = 246.0f;
  const float labelHeight = 18.0f;

  const float labelMarginX = 5.0f;
  const float labelMarginY = 5.0f;

  [self removeAllLabelLayers];

  int labelCount = 0;
  for  (const auto& item : newValues) {
    std::string label = item.second;
    const float value = item.first;
    const float originY = topMargin + ((labelHeight + labelMarginY) * labelCount);
    const int valuePercentage = (int)roundf(value * 100.0f);

    const float valueOriginX = leftMargin;
    NSString* valueText = [NSString stringWithFormat:@"%d%%", valuePercentage];

    [self addLabelLayerWithText:valueText
                        originX:valueOriginX
                        originY:originY
                          width:valueWidth
                         height:valueHeight
                      alignment:kCAAlignmentRight];

    const float labelOriginX = (leftMargin + valueWidth + labelMarginX);

    NSString *nsLabel = [NSString stringWithCString:label.c_str()
                                           encoding:[NSString defaultCStringEncoding]];
    [self addLabelLayerWithText:[nsLabel capitalizedString]
                        originX:labelOriginX
                        originY:originY
                          width:labelWidth
                         height:labelHeight
                      alignment:kCAAlignmentLeft];

    labelCount += 1;
    if (labelCount > 4) {
      break;
    }
  }
}

- (void)removeAllLabelLayers {
  for (CATextLayer* layer in labelLayers) {
    [layer removeFromSuperlayer];
  }
  [labelLayers removeAllObjects];
}

- (void)addLabelLayerWithText:(NSString*)text
                      originX:(float)originX
                      originY:(float)originY
                        width:(float)width
                       height:(float)height
                    alignment:(NSString*)alignment {
  CFTypeRef font = (CFTypeRef) @"Menlo-Regular";
  const float fontSize = 12.0;
  const float marginSizeX = 5.0f;
  const float marginSizeY = 2.0f;

  const CGRect backgroundBounds = CGRectMake(originX, originY, width, height);
  const CGRect textBounds = CGRectMake((originX + marginSizeX), (originY + marginSizeY),
                                       (width - (marginSizeX * 2)), (height - (marginSizeY * 2)));

  CATextLayer* background = [CATextLayer layer];
  [background setBackgroundColor:[UIColor blackColor].CGColor];
  [background setOpacity:0.5f];
  [background setFrame:backgroundBounds];
  background.cornerRadius = 5.0f;

  [[self.view layer] addSublayer:background];
  [labelLayers addObject:background];

  CATextLayer* layer = [CATextLayer layer];
  [layer setForegroundColor:[UIColor whiteColor].CGColor];
  [layer setFrame:textBounds];
  [layer setAlignmentMode:alignment];
  [layer setWrapped:YES];
  [layer setFont:font];
  [layer setFontSize:fontSize];
  layer.contentsScale = [[UIScreen mainScreen] scale];
  [layer setString:text];

  [[self.view layer] addSublayer:layer];
  [labelLayers addObject:layer];
}

@end
