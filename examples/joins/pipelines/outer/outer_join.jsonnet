////
// Template arguments:
//
// suffix : An arbitrary suffix appended to the name of this pipeline, for
//          disambiguation when multiple instances are created.
// tag : the tag of the docker image.
////
function(suffix, tag)
  [
  {
  "pipeline": {
    "name": "outer_join_"+suffix
  },
  "description": "A pipeline that lists all returns by zipcode joining stores and returns information.",
  "input": {
    "join": [
      {
        "pfs": {
          "repo": "stores",
          "branch": "master",
          "glob": "/STOREID(*).txt",
          "join_on": "$1",
          "outer_join": true
        }
      },
     {
       "pfs": {
         "repo": "returns",
         "branch": "master",
         "glob": "/*_STOREID(*).txt",
         "join_on": "$1",
         "outer_join": true
       }
     }
   ]
 },
 "transform": {
  "cmd": [ "python", "outer/main.py" ],
  "image": "pachyderm/example-joins-inner-outer:"+tag
  }
  },
  {
    "pipeline": {
      "name": "reduce_outer_"+suffix
    },
    "description": "A pipeline that consolidates all purchases by zipcode in one file.",
    "input": {
        "pfs": {
        "repo": "outer_join_"+suffix,
        "branch": "master",
        "glob": "/*"
        }
   },
   "transform": {
    "cmd": [ "bash" ],
    "stdin": [ "set -x", "FILES=/pfs/outer_join_"+suffix+"/*/*", "for f in $FILES", "do", "directory=`dirname $f`", "out=`basename $directory`",  "cat $f >> /pfs/out/${out}.txt", "done" ]
  }
  }
]