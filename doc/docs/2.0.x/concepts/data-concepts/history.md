# History

Pachyderm implements rich version-control and history semantics. This section
describes the core concepts and architecture of Pachyderm's version control
and the various ways to use the system to access historical data.

The following abstractions store the history of your data:

- **Commits**

    In Pachyderm, commits are the core version-control primitive that is
    similar to Git commits. [Commits](../commit/) represent an immutable snapshot of a
    filesystem and can be accessed with an ID. They have a parentage
    structure, where **new commits inherit content from their parents**. 
    You can think of this parentage structure as of a linked list or *a chain of commits*. 
 
- **Branches**

    [Branches](../branch/) are pointers to commits that are similar to Git branches. Typically,
    branches have semantically meaningful names such as `master` and `staging`.
    Branches are mutable, and they move along a growing chain of commits as you
    commit to the branch, and can even be reassigned to any commit within the
    repo by using the `pachctl create branch` command. The commit that a
    branch points to is referred to as the branches *head*, and the head's
    ancestors are referred to as *on the branch*. Branches can be substituted
    for commits in Pachyderm's API and behave as if the head of the branch
    were passed. 

## Ancestry Syntax

  - Pachyderm's commits and branches support a familiar Git syntax for
    referencing their history. A commit or branch parent can be referenced
    by adding a `^` to the end of the commit or branch. Similar to how
    `master` resolves to the head of `master`, `master^` resolves
    to the parent of the head. You can add multiple `^`s. For example,
    `master^^` resolves to the parent of the parent of the head
    `master`, and so on. Similarly, `master^3` has the same meaning as
    `master^^^`.

    Git supports two characters for ancestor references —`^` and `~`— with
    slightly different meanings. Pachyderm supports both characters as well,
    but their meaning is identical.

  - Also, Pachyderm supports a type of ancestor reference that Git does not:
    the **forward reference**, using the special character `.`. It
    resolves to commits on the beginning of commit chains. For example,
    `master.1` is the first (oldest) commit on the `master` branch, `master.2`
    is the second commit, and so on.

!!! Note
    Resolving ancestry syntax requires traversing chains of commits
    high numbers passed to `^` and low numbers passed to `.`. 
    These operations might take a long time.
    If you plan to repeatedly access an ancestor, you might want to resolve that
    ancestor with `pachctl inspect commit <repo>@<branch or commitID>`.

## View the Pipeline History

Pipelines are the main processing primitive in Pachyderm. However, they
expose version-control and history semantics similar to filesystem
objects. This is largely because, under the hood, they are implemented in
terms of filesystem objects. You can access previous versions of
a pipeline by using the same ancestry syntax that works for commits and
branches. For example, `pachctl inspect pipeline foo^` gives you the
previous version of the pipeline `foo`. The `pachctl inspect pipeline foo.1`
command returns the first ever version of that same pipeline. You can use
this syntax in all operations and scripts that accept pipeline names.

To view historical versions of a pipeline use the `--history`
flag with the `pachctl list pipeline` command:

```shell
pachctl list pipeline --history all
```
**System Response:**

```shell
NAME      VERSION INPUT     CREATED     STATE / LAST JOB
Pipeline2 1       input2:/* 4 hours ago running / success
Pipeline1 3       input1:/* 4 hours ago running / success
Pipeline1 2       input1:/* 4 hours ago running / success
Pipeline1 1       input1:/* 4 hours ago running / success
```
## View the Job History

Jobs do not have versioning semantics associated with them.
However, they are strongly associated with the pipelines that
created them. Therefore, they inherit some of their versioning
semantics. You can use the `-p <pipeline>` flag with the
`pachctl list job` command to list all the jobs that were run
for the latest version of the pipeline. 

Furthermore, you can get jobs from multiple versions of a
pipeline by passing the `--history` flag. For example,
`pachctl list job -p edges --history all` returns all jobs from all
versions of the pipeline edges.
