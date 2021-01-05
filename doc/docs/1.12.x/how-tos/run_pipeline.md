# Run a Pipeline on a Specific Commit

Sometimes you need to see the result of merging different commits
to analyze and identify correct combinations and potential flows in
data collection and processing. Or, you just want to rerun a failed job
manually.

Pachyderm enables you to run your pipelines with old commits or in different
handpicked combinations of old commits that are stored in separate
repositories or branches. This functionality is particularly useful for
cross-pipelines, but you can also use it with other types of pipelines.

For example, you have branches `A`, `B`, and `C`. Each of these branches has
had commits on it at various times, and each new commit
triggered a new job from the pipeline.

Your branches have the following commit history:

* Branch `A` has commits `A1`, `A2`, `A3`, and `A4`.
* Branch `B` has commits `B1` and `B2`.
* Branch `C` has commits `C1`, `C2`, and `C3`.

For example, you need to see the result of the
pipeline with the combination of data after `A4`, `B1`, and `C2` were
committed.
But none of the output commits were triggered on this particular combination.

To get the result of this combination, you can run the `pachctl run pipeline
cross-pipe` command.

!!! example
    ```shell
    pachctl run pipeline cross-pipe A4 B1 C2
    ```

This command triggers a new job that creates a commit on the
pipeline's output branch with the result of that combination of data.

Because `A4` is the head of branch `A`, you can also omit the `A4` commit
in the command and specify only the `C2` and `B1` commits:

!!! example
    ```
    pachctl run pipeline cross-pipe C2 B1
    ```

Pachyderm automatically uses the head for any branch that did not have a
commit specified. The order in which you specify the commits does not
matter. Pachyderm knows how to match them to the right place.

Also, you can reference the head commit of a branch by using the branch
name.

!!! example

    ```shell
    pachctl run pipeline cross-pipe A B1 C2
    ```

This behavior implies that if you want to re-run the pipeline on the
most recent commits, you can just run `pachctl run pipeline cross-pipe`.

If you try to use a commit from a branch that is not an input
branch, `pachctl` returns an error.
Similarly, specifying multiple commits from the same branch results in error.

You do not need to run the `pachctl run pipeline cross-pipe`
command to initiate a newly created pipeline. Pachyderm runs the new
pipelines automatically as you add new commits to the corresponding
input branches.
