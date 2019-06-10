# Running concurent pipelines

Sometimes you might decide to run the same pipeline on different branches and
you need to use the data from all the branches in different variations in
your pipeline. Such a pipeline is called cross pipeline or cross-pipe.

For example, you have branches `A`, `B`, and `C`. Each of these branches has
had commits on it at various times and each new commit
triggered a new job from the pipeline.

Your branches have the following commit history

* Branch `A` has commits `A1`, `A2`, `A3`, and `A4`.
* Branch `B` has commits `B1` and `B2`.
* Branch `C` has`C1`, `C2`, and `C3`.

For example you need to see the result of the
pipeline with the combination of data after `A4`, `B1`, and C2 were committed.
But none of the output commits were triggered on this particular combination.

To get the result of this combination, you can run the `pachctl run pipeline
cross-pipe` command.

**Example:**

```bash
pachctl run pipeline cross-pipe A4 B1 C2
```

This command triggerss a new job which creates a commit on the
pipeline's output branch with the result of that combination of data.

Because `A4` is the head of branch A, you can also run omit the `A4` commit
in the command and specify only the C2 and B1 commits:

```
pachctl run pipeline cross-pipe c2 b1
```

Pachyderm automatically uses the head for any branch that did not have a
commit specified. The order in which you specify the commits does not
matter. Pachyderm knows how to match them to the right place.

You can reference the head commit of a branch by using the branch
name.

**Example:**

```
pachctl run pipeline cross-pipe A B1 C2
```

This means that if you wanted to re-run the pipeline on the most recent
commits, you can just run `pachctl run pipeline cross-pipe`.

If you try to use a commit from a branch that is not an input
branch, `pachctl` returns an error.
Similarly, specifying multiple commits from the same branch results in error.

You do not need to run the `pachctl run pipeline cross-pipe`
command to initiate a newly created pipeline. Pachyderm runs the new
pipelines automatically as you add new commits to the corresponding
input branches.
