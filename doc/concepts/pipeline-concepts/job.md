# Job

A Pachyderm job is a unit of work that a pipeline triggers
when it detects new data in the input repository. Each
job runs your code against the current commit and
then submits the results to the output repository. A pipeline
triggers a new job every time you submit new changes into your
input source.

Each job has the following stages:

| Stage     | Description  |
| --------- | ------------ |
| Starting  | Pachyderm starts the job when it detects new data in the input repository. <br> The new data appears as a commit in the input repository, and Pachyderm <br> automatically launches the pipeline. Pachyderm spins the number of Pachyderm worker pods <br> specified in the pipeline spec and spreads the workload among them. |
| Running   | `pachd` runs the transformation code that is specified <br> in the pipeline specification against the newly committed data. |
| Merging   | Pachyderm concatenates the results of the processed <br> data into file or files and uploads them to the output repository. |
