#!/bin/bash
pachctl create-repo wordcount_input
pachctl start-commit wordcount_input -b master
pachctl put-file wordcount_input master -f README.md
pachctl finish-commit wordcount_input master
pachctl create-pipeline -f mapPipeline.json
