////
// Template arguments:
//
// suffix : An arbitrary suffix appended to the name of this pipeline, for
//          disambiguation when multiple instances are created.
////
function(suffix)
  [
  {
    "pipeline": {
      "name": "group_by_hospital_"+suffix
    },
    "description": "A pipeline that groups lab test results files by hospital using the files naming pattern.",
    "input": {
      "group": [
      {
          "pfs": {
            "repo": "labresults",
            "branch": "master",
            "glob": "/*-CLIA(*).txt",
            "group_by": "$1"
          }
      }
    ]
   },
   "transform": {
        "cmd": [ "bash" ],
        "stdin": ["mkdir -p /pfs/out/${PACH_DATUM_labresults_GROUP_BY}/","cp /pfs/labresults/* /pfs/out/${PACH_DATUM_labresults_GROUP_BY}/"]
        }
  }
,
 {
  "pipeline": {
    "name": "reduce_group_by_hospital_"+suffix
  },
  "description": "A pipeline that consolidates all lab results by hospital in one file.",
  "input": {
      "pfs": {
      "repo": "group_by_hospital_"+suffix,
      "branch": "master",
      "glob": "/*"
      }
 },
 "transform": {
  "cmd": [ "bash" ],
  "stdin": [ "set -x", "FILES=/pfs/group_by_hospital_"+suffix+"/*/*", "for f in $FILES", "do", "directory=`dirname $f`", "out=`basename $directory`",  "cat $f >> /pfs/out/${out}.txt", "done" ]
}
}
]