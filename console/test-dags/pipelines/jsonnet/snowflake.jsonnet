/*
title: Snowflake Integration
description: "Creates a cron pipeline that can execute a query against a Snowflake database and return the results in a single output file."
args:
- name: name
  description: The name of the pipeline.
  type: string
- name: inputrepo
  description: The suffix to use on the cron pipeline's input repo.
  type: string
  default: tick
- name: spec
  description: The cron spec to use. 
  type: string
  default: "@never"
- name: secret
  description: "The name of the k8s secret containing the Snowflake connection details. The secret itself must contain three values: snow-account, snow-user, and snow-password."
  type: string
  default: snowflake-connection
- name: outputformat
  description: The format for the output file.
  type: string
  default: csv
- name: outputfile
  description: 'The name of the output file to produce. We assume that the root is \"/pfs/out/\".'
  type: string
  default: output.csv
- name: query
  description: The query to execute.
  type: string
*/

function(name, 
         inputrepo="tick", 
         spec="@never", 
         overwrite=true, 
         secret="snowflake-connection", 
         outputformat="csv",
         outputfile="output.csv",
         headers=false,
         query)
{
    "pipeline": {
      "name": name
    },
    "input": {
      "cron": {
          "name": inputrepo,
          "spec": spec,
          "overwrite": overwrite
      }
    },
    "transform": {
      "secrets": [ 
          {   "name": secret,
              "key": "snow-account",
              "envVar": "SNOWSQL_ACCOUNT"
          },
          {   "name": secret,
              "key": "snow-user",
              "envVar": "SNOWSQL_USER"
          },
          {   "name": secret,
              "key": "snow-password",
              "envVar": "SNOWSQL_PWD"
          }
       ],
       
      "cmd":["sh"],
      "stdin": ["./run_snowsql.sh " + outputformat + " /pfs/out/" + outputfile + " " + headers + " '" + query + "'"],
      "image": "dgeorg42/mldm-snowflake-integration:v1.0.0"
    },
    "autoscaling": true,
    "datumTries": 1
}