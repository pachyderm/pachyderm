# Job

A Pachyderm job is an execution of a pipeline that triggers
when new data is detected in an input repository. Each
job runs your code against the current commit and
then submits the results to the output repository and creates a single output commit. A pipeline
triggers a new job every time you submit new changes, a commit, into your
input source.

Each job has the following stages:

| Stage     | Description  |
| --------- | ------------ |
| Starting  | Pachyderm starts the job when it detects new data in the input repository. <br> The new data appears as a commit in the input repository, and Pachyderm <br> automatically launches the job. Pachyderm spins the number of Pachyderm worker pods <br> specified in the pipeline spec and spreads the workload among them. |
| Running   | Pachyderm runs the transformation code that is specified <br> in the pipeline specification against the data in the input commit. |
| Merging   | Pachyderm concatenates the results of the processed <br> data into one or more files, uploads them to the output repository, completes the final output commits, and creates/persists all the versioning metadata |
