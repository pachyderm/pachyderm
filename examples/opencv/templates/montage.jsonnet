////
// Template arguments:
//
// suffix : An arbitrary suffix appended to the name of this pipeline, for
//          disambiguation when multiple instances are created.
// left : the first repo/branch from which this pipeline will read images
//        to be placed in the montage (must be of the form '<repo>' or
//        '<repo>@<branch>'. <branch> defaults to 'master' if unset.
// right : the second repo/branch from which this pipeline will read images
//         to be placed in the montage (must be of the form '<repo>' or
//         '<repo>@<branch>'. <branch> defaults to 'master' if unset.
////
function(suffix, left, right)

// validate arguments
assert std.length(suffix) > 0 : "Must set suffix to a nonempty string";
assert std.length(left) > 0 : "Must set src to 'repo' or 'repo@branch'";
assert std.length(right) > 0 : "Must set src to 'repo' or 'repo@branch'";

// split 'left' into 'repo' and 'branch'
local left_repobranch = std.splitLimit(left, "@", 1);
local left_repo = left_repobranch[0];
local left_branch = if std.length(left_repobranch) > 1
               then left_repobranch[1]
               else "master";

// split 'right' into 'repo' and 'branch'
local right_repobranch = std.splitLimit(right, "@", 1);
local right_repo = right_repobranch[0];
local right_branch = if std.length(right_repobranch) > 1
               then right_repobranch[1]
               else "master";

// tmp variables for the pipeline description
local l_full = left_repo + "@" + left_branch;
local r_full = right_repo + "@" + right_branch;

// result pipeline spec:
{
  pipeline: { name: "montage-"+suffix },
  description: "A pipeline that combines images from "+l_full+" and "+r_full+" into a montage.",
  input: {
    cross: [
      {
        pfs: {
          name: "left",
          glob: "/",
          repo: left_repo,
          branch: left_branch,
        }
      },
      {
        pfs: {
          name: "right",
          glob: "/",
          repo: right_repo,
          branch: right_branch,
        }
      }
    ]
  },
  transform: {
    cmd: [ "sh" ],
    image: "dpokidov/imagemagick:7.0.10-58",
    stdin: [ "montage -shadow -background SkyBlue -geometry 300x300+2+2 /pfs/left/* /pfs/right/* /pfs/out/montage.png" ]
  }
}

