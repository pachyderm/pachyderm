# Job

A Pachyderm job is a unit of work, that pipeline triggers
when it detects a new commit in the input repository. Each
job runs your code against the current state of the and
then submits the results to the output repository. The pipeline
triggers a new job every time you submit your changes to your
input source.


