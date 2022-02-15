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
      "name": "group_store_revenue_"+suffix
    },
    "description": "A pipeline that groups purchases and returns by storeId to calculate the gross_revenue minus returns of each store.",
    "input": {
      "group": [
          {
            "pfs": {
              "repo": "stores",
              "branch": "master",
              "glob": "/STOREID(*).txt",
              "group_by": "$1"
            }
          },
         {
           "pfs": {
             "repo": "purchases",
             "branch": "master",
             "glob": "/*_STOREID(*).txt",
             "group_by": "$1"
           }
         },
         {
          "pfs": {
            "repo": "returns",
            "branch": "master",
            "glob": "/*_STOREID(*).txt",
            "group_by": "$1"
          }
        }
    ]
   },
   "transform": {
    "cmd": [ "python", "main.py" ],
    "image": "pachyderm/example-group:"+tag
  }
}
 

]