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
  This command shows the progress of all global jobs, in an aggregated format. 
  
  For each commit (global job), Pachyderm shows the number of subjobs run, and  a progress bar that mirroring all sub-jobs in each commit's DAG.

  **Example:**
  ```shell
  ID                               SUBJOBS PROGRESS CREATED     MODIFIED
  8862dae3ee1348ba8492b5863c32cae5 2       ▇▇▇▇▇▇▇▇ 3 hours ago 3 hours ago 
  20f5ac9d632e47fb9b8e763f8bdce178 1       ▇▇▇▇▇▇▇▇ 3 hours ago 3 hours ago 
  4b9aaebf49384842a72537d7e1532a63 1       ▇▇▇▇▇▇▇▇ 3 hours ago 3 hours ago 
  ```

  The progress bar is equally divided to the number of steps, or pipelines,
  you have in your DAG. In the example above, it is two steps.
  If one of the sub-jobs fails, you will see the progress bar turn red
  for that pipeline step. To troubleshoot, look into that particular
  pipeline execution.
## `pachctl list job <global-id>`
  This command shows the status of all the pipelines executions that run in the context of this global job.

  For each pipeline executed as part of this global job, Pachyderm shows the time since each sub-job started and its duration, the number of datums in the **PROGRESS** section,  and other information.
  The format of the progress column is `DATUMS PROCESSED + DATUMS SKIPPED / TOTAL DATUMS`.

  For more information, see
  [Datum Processing States](../../../concepts/pipeline-concepts/datum/datum-processing-states/).

  **Example:**
  ```shell
  % pachctl list job 8862dae3ee1348ba8492b5863c32cae5
  ID                               PIPELINE STARTED     DURATION  RESTART PROGRESS  DL       UL       STATE   
  8862dae3ee1348ba8492b5863c32cae5 edges    3 hours ago 1 second  0       1 + 1 / 2 78.7KiB  37.15KiB success 
  8862dae3ee1348ba8492b5863c32cae5 montage  3 hours ago 3 seconds 0       1 + 0 / 1 195.3KiB 815.1KiB success 
  ```

.......

## `pachctl list job -p <pipeline>`
  This command shows the status of a pipeline's executions across all commits.

  For each commit, Pachyderm shows the time the job started with its duration, data downloaded and uploaded, STATE of the pipeline execution, the number of datums in the **PROGRESS** section,  and other information.
  The format of the progress column is `DATUMS PROCESSED + DATUMS SKIPPED / TOTAL DATUMS`.

  For more information, see
  [Datum Processing States](../../../concepts/pipeline-concepts/datum/datum-processing-states/).

  **Example:** 
  ```shell
  % pachctl list job -p edges
  ID                               PIPELINE STARTED     DURATION RESTART PROGRESS  DL       UL       STATE   
  8862dae3ee1348ba8492b5863c32cae5 edges    3 hours ago 1 second 0       1 + 1 / 2 78.7KiB  37.15KiB success 
  4b9aaebf49384842a72537d7e1532a63 edges    3 hours ago 1 second 0       1 + 0 / 1 57.27KiB 22.22KiB success 
  ```

!!! note "See Also"
    [Pipeline Troubleshooting](../../../troubleshooting/pipeline_troubleshooting/)
