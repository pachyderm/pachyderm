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
| Starting  | Pachyderm starts the job when it detects new data in the input repository. <br> A job is in the *starting* state when it is pulling the user Docker image, creating <br> the worker pods, and waiting for Kubernetes to schedule them. |
| Running   | Pachyderm is running the user code over all <br> input datums. |
| Merging   | Pachyderm concatenates the results of the processed <br> data into one or more files, uploads them to the output repository, completes <br> the final output commits, and creates/persists all the versioning metadata. |
