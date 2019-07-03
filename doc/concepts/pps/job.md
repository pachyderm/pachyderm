# Job

A Pachyderm job is a unit of work that a pipeline triggers
when it detects new data in the input repository. Each
job runs your code against the current state of the and
then submits the results to the output repository. The pipeline
triggers a new job every time you submit your changes to your
input source.

Each job has the following stages:

| Stage     | Description  |
| --------- | ------------ |
| Starting  | Pachyderm starts the job when it detects the new data in the input repository. <br> The new data appears as a commit and Pachyderm automatically starts the processing <br>. Depending on the pipeline specification, Pachyderm spins one more Pachyderm worker <br> pods to process the data and spreads the workload amongst them. |
| Running   | `pachd` runs the transformation code that is specified <br> in the pipeline specification against the newly committed data. |
| Merging   | Pachyderm concatenates the results of the processed <br> data into file or files and uploads them to the output repository. |
