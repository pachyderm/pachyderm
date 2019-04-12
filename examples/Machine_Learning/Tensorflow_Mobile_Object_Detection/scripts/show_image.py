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

from IPython.display import Image, HTML, display

    
root = "tf_files/flower_photos/"
with open(root+"/LICENSE.txt") as f:
    attributions = f.readlines()[4:]
attributions = [line.split(' CC-BY') for line in attributions]
attributions = dict(attributions)
    
def show_image(image_path):
    display(Image(image_path))
    
    image_rel = image_path.replace(root,'')
    caption = "Image " + ' - '.join(attributions[image_rel].split(' - ')[:-1])
    display(HTML("<div>%s</div>" % caption))
