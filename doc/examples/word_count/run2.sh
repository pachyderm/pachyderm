#!/bin/bash
pachctl stop-pipeline wordcount-map
pachctl start-commit wordcount_input master
pachctl put-file wordcount_input master -f README.md
pachctl finish-commit wordcount_input master
pachctl start-commit wordcount_input master
pachctl put-file wordcount_input master -f README.md
pachctl finish-commit wordcount_input master
#pachctl start-pipeline wordcount-map
