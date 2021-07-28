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
|CREATED| An output commit exists, but the job has not been started by a worker yet.|
|STARTING| The worker has allocated resources for the job (that is, the job counts towards parallelism), but it is still waiting on the inputs to be ready.|
|RUNNING|The worker is processing datums.|
|EGRESS|The worker has completed all the datums but is uploading the output to the egress endpoint.|
|FAILURE|The worker encountered too many errors when processing a datum.|
|KILLED|The job timed out, the output commits were deleted, or a user otherwise called StopJob|
|SUCCESS| None of the bad stuff happened.|


