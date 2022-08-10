# Branch

A Pachyderm branch is a pointer to a commit that
moves along with new commits as they are submitted. By default,
when you create a repository, Pachyderm does not create any branches.
Most users prefer to create a `master` branch by initiating the first
commit and specifying the `master` branch in the `put file` command.

Branches enable collaboration between teams of data scientists.
However, many users find it sufficient to
use the master branch for all their work. Although the concept of a
branch is similar to Git branches, in most cases, branches are not
used as extensively as in source code version-control systems.

A branch also stores information about [provenance](provenance.md), the other
branches it uses as input and which rely on its output. These branch relationships
are how Pachyderm knows which data each pipeline relies on, and which branches and
repos should be included in each commit.

Each branch has a `HEAD` which references the latest commit in the
branch. Pachyderm pipelines look at the `HEAD` of the branch
for changes and, if they detect new changes, trigger a job. **When you
commit a new change, the `HEAD` of the branch moves to the latest commit.**

You can create additional branches to experiment with the data (`pachctl create branch <myrepo>@<branchname>`. Optionally, you can add `--head  <myrepo>@<master>` for the head of the new branch to reference the head commit on master).

To view a list of branches in a repo, run the `pachctl list branch <myrepo>` command.

!!! example
    ```shell
    pachctl list branch images
    ```

    **System Response:**

    ```shell
    BRANCH HEAD
    master c32879ae0e6f4b629a43429b7ec10ccc
    ```
!!! Attention "It is important to note that..."
    - Deleting a branch (`pachctl delete branch <myrepo>@<branchname>`) does not delete the commits on it.
    - All branches must have a head commit. 