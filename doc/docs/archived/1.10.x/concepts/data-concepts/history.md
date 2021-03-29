# History

Pachyderm implements rich version-control and history semantics. This section
describes the core concepts and architecture of Pachyderm's version control
and the various ways to use the system to access historical data.

The following abstractions store the history of your data:

* **Commits**

  In Pachyderm, commits are the core version-control primitive that is
  similar to Git commits. Commits represent an immutable snapshot of a
  filesystem and can be accessed with an ID. Commits have a parentage
  structure, where new commits inherit content from their parents.
  You can think of this parentage structure as of a linked list or *a chain of
  commits*. Commit IDs are useful if you want to have a static pointer to
  a snapshot of a filesystem. However, because they are static, their use is
  limited. Instead, you mostly work with branches.

* **Branches**

  Branches are pointers to commits that are similar to Git branches. Typically,
  branches have semantically meaningful names such as `master` and `staging`.
  Branches are mutable, and they move along a growing chain of commits as you
  commit to the branch, and can even be reassigned to any commit within the
  repo by using the `pachctl create branch` command. The commit that a
  branch points to is referred to as the branches *head*, and the head's
  ancestors are referred to as *on the branch*. Branches can be substituted
  for commits in Pachyderm's API and behave as if the head of the branch
  were passed. This allows you to deal with semantically meaningful names
  for commits that can be updated, rather than static opaque identifiers.

## Ancestry Syntax

Pachyderm's commits and branches support a familiar Git syntax for
referencing their history. A commit or branch parent can be referenced
by adding a `^` to the end of the commit or branch. Similar to how
`master` resolves to the head commit of `master`, `master^` resolves
to the parent of the head commit. You can add multiple `^`s. For example,
`master^^` resolves to the parent of the parent of the head commit of
`master`, and so on. Similarly, `master^3` has the same meaning as
`master^^^`.

Git supports two characters for ancestor references —`^` and `~`— with
slightly different meanings. Pachyderm supports both characters as well,
but their meaning is identical.

Also, Pachyderm supports a type of ancestor reference that Git does not—
forward references, these use a different special character `.` and
resolve to commits on the beginning of commit chains. For example,
`master.1` is the first (oldest) commit on the `master` branch, `master.2`
is the second commit, and so on.

Resolving ancestry syntax requires traversing chains of commits
high numbers passed to `^` and low numbers passed to `.`. These operations
require traversing a large number of commits which might take a long time.
If you plan to repeatedly access an ancestor, you might want to resolve that
ancestor to a static commit ID with `pachctl inspect commit` and use
that ID for future accesses.

## View the Filesystem Object History

Pachyderm enables you to view the history of filesystem objects by using
the `--history` flag with the `pachctl list file` command. This flag
takes a single argument, an integer, which indicates how many historical
versions you want to display. For example, you can get
the two most recent versions of a file with the following command:

```shell
pachctl list file repo@master:/file --history 2
```

**System Response:**

```shell
COMMIT                           NAME  TYPE COMMITTED      SIZE
73ba56144be94f5bad1ce64e6b96eade /file file 16 seconds ago 8B
c5026f053a7f482fbd719dadecec8f89 /file file 21 seconds ago 4B
```

This command might return a different result from if you run
`pachctl list file repo@master:/file` followed by `pachctl list file
repo@master^:/file`. The history flag looks for changes
to the file, and the file might not be changed with every commit.
Similar to the ancestry syntax above, because the history flag requires
traversing through a linked list of commits, this operation can be
expensive. You can get back the full history of a file by passing
`all` to the history flag.

**Example:**

```shell
pachctl list file edges@master:liberty.png --history all
```

**System Response:**

```shell
COMMIT                           NAME         TYPE COMMITTED    SIZE
ff479f3a639344daa9474e729619d258 /liberty.png file 23 hours ago 22.22KiB
```

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

A common operation with pipelines is reverting a pipeline to a previous
version.
To revert a pipeline to a previous version, run the following command:

```shell
pachctl extract pipeline pipeline^ | pachctl create pipeline
```

## View the Job History

Jobs do not have versioning semantics associated with them.
However, they are strongly associated with the pipelines that
created them. Therefore, they inherit some of their versioning
semantics. You can use the `-p <pipeline>` flag with the
`pachctl list job` command to list all the jobs that were run
for the latest version of the pipeline. To view a previous version
of a pipeline you can add the caret symbol to the end of the
pipeline name. For example `-p edges^`.

Furthermore, you can get jobs from multiple versions of
pipelines by passing the `--history` flag. For example,
`pachctl list job  --history all` returns all jobs from all
versions of all pipelines.

To view job history, run the following command:

* By using the `-p` flag:

  ```shell
  pachctl list job -p <pipeline^>
  ```

* By using the `history` flag:

  ```shell
  pachctl list job --history all
  ```
