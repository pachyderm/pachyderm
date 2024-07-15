/*
title: Edges From Gist
description: "Simple example pipeline github Gist."
args:
- name: suffix
  description: Pipeline name suffix
  default: 1
- description: Input repo to pipeline.
  type: string
  default: images
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