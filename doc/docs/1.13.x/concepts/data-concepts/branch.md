# Branch

A Pachyderm branch is a pointer, or an alias, to a commit that
moves along with new commits as they are submitted. By default,
when you create a repository, Pachyderm does not create any branches.
Most users prefer to create a `master` branch by initiating the first
commit and specifying the `master` branch in the `put file` command.
Also, you can create additional branches to experiment with the data.
Branches enable collaboration between teams of data scientists.
However, many users find it sufficient to
use the master branch for all their work. Although the concept of a
branch is similar to Git branches, in most cases, branches are not
used as extensively as in source code version-control systems.

Each branch has a `HEAD` which references the latest commit in the
branch. Pachyderm pipelines look at the `HEAD` of the branch
for changes and, if they detect new changes, trigger a job. When you
commit a new change, the `HEAD` of the branch moves to the latest commit.

Sometimes you might see no `HEAD` in a branch. This happens when a branch
does not have any commits yet or if all commits were deleted.
This can also occur if you create a pipeline that takes an
input that does not yet exist. In the example below, the `test` branch
does not have a `HEAD`.

To view a list of branches in a repo, run the `pachctl list branch` command.

!!! example
    ```shell
    pachctl list branch images
    ```

    **System Response:**

    ```shell
    BRANCH HEAD
    test   -
    master c32879ae0e6f4b629a43429b7ec10ccc
    ```
