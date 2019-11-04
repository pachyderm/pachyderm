# Skip Failed Datums

!!! note "Summary"
    The `err_cmd` parameter enables you to fail a datum without failing the
    whole job.

!!! note
    Before you read this section, make sure that you understand such
    concepts as [Datum](../../concepts/pipeline-concepts/datum/) and
    [Pipeline](../../concepts/pipeline-concepts/pipeline/).

When Pachyderm processes your data, it breaks it up into units of
computation called datums. Each datum is processed separately.
In a basic pipeline configuration, a failed datum results in a failed
pipeline. However, in some cases, you might not need all datums
to consider a job successful. If a datum can be further processed in
your downstream pipelines without impacting the result, Pachyderm can
mark that datum as *recovered* and mark the job as successful.

To configure a condition under which you want your failed datums not
to fail the whole job, you can add your custom error code in
`err_cmd` and `err_stdin` fields in your pipeline specification.

For example, your DAG consists of two pipelines:

* The pipeline 1 cleans the data.
* The pipeline 2 trains your model by using the data from the first pipeline.

That means that the second pipeline takes the results of the first pipeline
from its output repository and uses that data to train a model. In some cases,
you might not need all the datums in the first pipeline to be successful
to run the second pipeline.

The following diagram describes how Pachyderm transformation and error
code work:

![err_cmd logic](../assets/images/err_cmd_workflow.svg)

Here is what is happening in the diagram above:

1. Pachyderm executes the transformation code that you defined in
the `cmd` field against your datums.
1. If a datum is processed without errors, Pachyderm marks it as
`processed`.
1. If a datum fails, Pachyderm marks it as *skipped* and executes your
error code (`err_cmd`) on that datum.
1. If the code in `err_cmd` successfully runs on the *skipped* datum,
Pachyderm marks the skipped datum as `recovered` and marks the job as
successful.
1. If the `err_cmd` code fails on the skipped datum, the datum is marked
as failed, and, consequently, the job is marked as failed.

You can view the processed, skipped, and recovered datums in the `PROGRESS`
field in the output of the `pachctl list job` command:

![datums in progress](../assets/images/datums_in_progress.svg)

Only processed datums are used in downstream pipelines if there are any.
For example, in your first pipeline, Pachyderm processes three datums.
If one of the datums is marked as *recovered* and two others are
successfully processed, only these two successful datums are used in
the next pipeline.

!!! note "See also"
    [Example err_cmd pipeline](https://github.com/pachyderm/pachyderm/tree/master/examples/err-cmd-example/)

