# Monitor Job Progress

After a pipeline starts a job, you can run one of the following commands
to monitor its status:

* `pachctl list pipeline`

  This command shows all pipelines that run in your cluster
  and their current status. Jobs for each pipeline will finish with either
  `success` or `fail` status.

  **Example:**

  ```bash
  NAME    VERSION INPUT                 CREATED       STATE / LAST JOB    DESCRIPTION
  montage 1       (edges:/ ⨯ images:/)  2 seconds ago starting / starting A montage pipeline
  edges   1       images:/*             2 seconds ago running / starting  An edge detection pipeline.
  ```

* `pachctl list job`

  This command shows jobs that were run for each pipeline. For each job,
  Pachyderm shows the number of datums in the **PROGRESS** section, the amount
  of downloaded and uploaded data, duration, and other important information.

  **Example:**

  ```bash
  svetlanakarslioglu@Svetlanas-MBP examples % pachctl list job
  ID                               PIPELINE STARTED       DURATION           RESTART PROGRESS    DL       UL       STATE
  7321952b9a214d3dbb64cc4369cc67da montage  6 minutes ago 1 second           0       1 + 0 / 1   371.9KiB 1.283MiB success
  95adc138e82e48949909364e8b9dbb53 edges    6 minutes ago 1 second           0       2 + 1 / 3   181.1KiB 111.4KiB success
  84fe22432f22492c9fd4f23036c3c8b5 montage  6 minutes ago Less than a second 0       1 + 0 / 1   79.49KiB 378.6KiB success
  2fbbc54ab3514d8a94d1b7a75bab96a7 edges    6 minutes ago Less than a second 0       1 + 0 / 1   57.27KiB 22.22KiB success
  ```

* `pachctl list commit <first-repo-in-dag>`

  This command, when ran against the first repo in your DAG,
  shows the status of the downstream commits further in the DAG.
  In the [Hyperparameter Tuning example](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter), we have four pipelines,
  or a four-stage pipeline. Every consequent pipeline takes the results
  in the output repository of the previous pipeline and performs a
  computation. Therefore, each step is executed one after another.
  The **PROGRESS** bar in the output of the `pachctl list commit <first-repo-in-dag>`
  command reflects these changes.

  The following animation shows how the progress bar is updated
  when a job for each pipeline completes.

  <p><small>(Click to enlarge)</small></p>
  [ ![Progress bar](../../assets/images/list_commit_progress_bar.gif)](../../assets/images/list_commit_progress_bar.gif)
