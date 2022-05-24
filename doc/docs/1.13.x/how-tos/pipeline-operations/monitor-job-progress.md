# Monitor Job Progress

After a pipeline starts a job, you can run one of the following commands
to monitor its status:

## `pachctl list pipeline`

  This command shows all the pipelines that run in your cluster
  and the status of the last job. 

  - The `STATE` column shows the current
    state of the pipeline. If you see that a pipeline is in `running`
    state, it means that pods were spun up in the underlying Kubernetes
    cluster for this pipeline. The running state does not necessarily mean
    that the pipeline is actively processing a job. If you see `failed` in
    the `STATE`
    column, this means that the Kubernetes cluster failed to schedule a pod for
    this pipeline.

  - The `LAST JOB` column shows the status of the most recent job that ran
    for this pipeline, which can be either `success`, `failed`, or
    `crashing`. If a pipeline is in a `failed` state, you need to find the
    reason of the failure and fix it. The crashing state indicates that
    the pipeline worker is failing for potentially transient reasons. The
    most common reasons for crashing are **image pull failures**, such as
    incorrect image name or registry credentials, or scheduling failures,
    such as not enough resources on your Kubernetes cluster.

  **Example:**

  ```shell
  NAME    VERSION INPUT                 CREATED       STATE / LAST JOB    DESCRIPTION
  montage 1       (edges:/ ⨯ images:/)  2 seconds ago starting / starting A montage pipeline
  edges   1       images:/*             2 seconds ago running / starting  An edge detection pipeline.
  ```

## `pachctl list job`

  This command shows the jobs that were run for each pipeline. 
  
  For each job, Pachyderm shows the related pipeline, the time since the job started and its duration, the number of datums in the **PROGRESS** section,  and other information.
  The format of the progress column is `DATUMS PROCESSED + DATUMS SKIPPED / TOTAL DATUMS`.

  For more information, see
  [Datum Processing States](../../../concepts/pipeline-concepts/datum/datum-processing-states/).

  **Example:**

  ```shell
  % pachctl list job
  ID                               PIPELINE STARTED       DURATION           RESTART PROGRESS    DL       UL       STATE
  7321952b9a214d3dbb64cc4369cc67da montage  6 minutes ago 1 second           0       1 + 0 / 1   371.9KiB 1.283MiB success
  95adc138e82e48949909364e8b9dbb53 edges    6 minutes ago 1 second           0       2 + 1 / 3   181.1KiB 111.4KiB success
  84fe22432f22492c9fd4f23036c3c8b5 montage  6 minutes ago Less than a second 0       1 + 0 / 1   79.49KiB 378.6KiB success
  2fbbc54ab3514d8a94d1b7a75bab96a7 edges    6 minutes ago Less than a second 0       1 + 0 / 1   57.27KiB 22.22KiB success
  ```

## `pachctl list commit <repo>`

  This command shows the status of the downstream jobs further in
  the DAG that result from this commit.
  
  In the [Hyperparameter Tuning example](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/hyperparameter), we have four pipelines,
  or a four-stage pipeline. Every subsequent pipeline takes the results
  in the output repository of the previous pipeline and performs a
  computation. Therefore, each step is executed one after another.
  The **PROGRESS** bar in the output of the `pachctl list commit <first-repo-in-dag>`
  command reflects these changes.

  Running the command against the first repo in the DAG displays
  a progress bar that shows job progress for all steps in your DAG.

  The following animation shows how the progress bar is updated
  when a job for each pipeline completes.

  ![Progress bar](../../../assets/images/list_commit_progress_bar.gif)

  The progress bar is equally divided to the number of steps, or pipelines,
  you have in your DAG. In the example above, it is four steps.
  If one of the jobs fails, you will see the progress bar turn red
  for that pipeline step. To troubleshoot, look into that particular
  pipeline job.

!!! note "See Also"
    [Pipeline Troubleshooting](../../../troubleshooting/pipeline-troubleshooting/)
