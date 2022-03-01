# Skip Failed Datums

!!! note "TL;DR"
    The `err_cmd` parameter enables you to fail a datum without failing the
    whole job.

!!! note
    Before you read this section, make sure that you understand such
    concepts as [Datum](../../../concepts/pipeline-concepts/datum/) and
    [Pipeline](../../../concepts/pipeline-concepts/pipeline/).

When Pachyderm processes your data, it breaks it up into units of
computation called datums. Each datum is processed separately.
In a basic pipeline configuration, a failed datum results in a failed
job. However, in some cases, you might not need all datums
to consider a job successful. If your downstream pipelines can be run
on only the successful datums instead of needing all the datums to be
successful, Pachyderm can mark some datums as *recovered* which means
that they failed with a non-critical error, but the successful datums
will be processed.

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

![err_cmd logic](../../assets/images/err_cmd_workflow.svg)

Here is what is happening in the diagram above:

1. Pachyderm executes the transformation code that you defined in
the `cmd` field against your datums.
1. If a datum is processed without errors, Pachyderm marks it as
`processed`.
1. If a datum fails, Pachyderm marks executes your
error code (`err_cmd`) on that datum.
1. If the code in `err_cmd` successfully runs on the *skipped* datum,
Pachyderm marks the skipped datum as `recovered`. The datum is in a
failed state and, therefore, the pipeline does not put it into the output
repository, but successful datums continue onto the next step in your DAG.
1. If the `err_cmd` code fails on the skipped datum, the datum is marked
as failed, and, consequently, the job is marked as failed.

You can view the processed, skipped, and recovered datums in the `PROGRESS`
field in the output of the `pachctl list job` command:

![Datums in progress](../../assets/images/datums_in_progress.svg)

Pachyderm writes only processed datums of successful jobs to the output
commit so that these datums can be processed by downstream pipelines.
For example, in your first pipeline, Pachyderm processes three datums.
If one of the datums is marked as *recovered* and two others are
successfully processed, only these two successful datums are used in
the next pipeline.

If you want to let the job proceed with only the successful datums being
written to the output, set `"err_cmd" : ["true"]`. The failed datums,
which are "recovered" by `err_cmd` in this way, will be retried on
the next job, just as failed datums.

!!! note "See Also:"
    [Example err_cmd pipeline](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/err_cmd/)
