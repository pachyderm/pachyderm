/*
title: Image edges
description: "Simple example pipeline."
args:
- name: suffix
  description: Pipeline name suffix
  type: string
  default: 1
- name: src
  description: Input repo to pipeline.
  type: string
  default: test
*/
function(suffix, src)
{
  pipeline: { name: "edges-"+suffix },
  description: "OpenCV edge detection on "+src,
  input: {
    pfs: {
      name: "images",
      glob: "/*",
      repo: src,
    }
  },
  transform: {
    cmd: [ "python3", "/edges.py" ],
    image: "pachyderm/opencv:0.0.1"
  }
}