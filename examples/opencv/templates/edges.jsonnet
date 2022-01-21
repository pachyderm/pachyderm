////
// Template arguments:
//
// suffix : An arbitrary suffix appended to the name of this pipeline, for
//          disambiguation when multiple instances are created.
// src : the repo from which this pipeline will read the images to which
//       it applies edge detection.
////
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