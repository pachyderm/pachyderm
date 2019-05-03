# History

Pachyderm implements rich version-control and history semantics. This doc
lays out the core concepts and architecture of Pachyderm's version-control
and the various ways to use the system to access historical data.

## Commits

Commits are the core version-control primitive in Pachyderm, similar to
git, commits represent an immutable snapshot of a filesystem and can be
accessed with an ID. Commits have a parentage structure, with new commits
inheriting content from their parents and then adding to it, you can think
of it as a linked list, it's also often referred to as "a chain of
commits." Commit IDs are useful if you want to have a static pointer to
a snapshot of a filesystem. However, because they're static their use is
limited, and you'll mostly deal with branches instead.

## Branches

Branches are tags that point to commits, again similar to git, branches
have semantically meaningful names such as `master` and `staging`.
Branches are mutable, they move along a growing chain of commits as you
commit to the branch, and can even be reassigned to any commit within the
repo (with `create branch`). The commit that a branch points to is
referred to as the branches "head," and the head's ancestors are referred
to as "on the branch." Branches can be substituted for commits in
Pachyderm's API and will behave as if the head of the branch were passed.
This allows you to deal with semantic meaningful names for commits that
can be updated, rather than static opaque identifiers.

## Ancestry Syntax

Pachyderm's commits and branches support a familiar git syntax for
referencing their history. A commit or branch's parent can be referenced
by adding a `^` to the end of the commit or branch. Similar to how
`master` will resolves to the head commit of `master`, `master^` resolves
to the parent of the head commit. You can add multiple `^`s, for example
`master^^` resolves to the parent of the parent of the head commit of
`master`, and so on. This gets unwieldy quickly so it can also be written
as `master^3`, which has the same meaning as `master^^^`. Git supports two
characters for ancestor references, `^` and `~` with slightly different
meanings, Pachyderm supports both characters as well, for familiarity
sake, but their meaning is identical.

Pachyderm also supports a type of ancestor reference that git doesn't:
forward references, these use a different special character `.` and
resolve to commits on the beginning of commit chains. For example
`master.1` is the first (oldest) commit on the `master` branch, `master.2`
is the second commit, and so on.

Note that resolving ancestry syntax requires traversing chains of commits
high numbers passed to `^` and low numbers passed to `.` will require
traversing a large number of commits which, will take a long time. If you
plan to repeatedly access an ancestor it may be worth it to resolve that
ancestor to a static commit ID with `inspect commit` and use that ID for
future accesses.

## The history flag

Pachyderm also allows you to list the history of objects using
a `--history` flag. This flag takes a single argument, an integer, which
indicates how many historical versions you want. For example you can get
the two most recent versions of a file with the following command:

```sh
$ pachctl list file repo@master:/file --history 2
COMMIT                           NAME  TYPE COMMITTED      SIZE
73ba56144be94f5bad1ce64e6b96eade /file file 16 seconds ago 8B
c5026f053a7f482fbd719dadecec8f89 /file file 21 seconds ago 4B
```

Note that this isn't necessarily the same result you'd get if you did:
`pachctl list file repo@master:/file` followed by `pachctl list file
repo@master^:/file`, because the history flag actually looks for changes
to the file, and the file need not change in every commit. Similar to the
ancestry syntax above, the history flag requires traversing through
a linked list of commits, and thus can be expensive. You can get back the
full history of a file by passing `-1` to the history flag.


## Pipelines

Pipelines are the main processing primitive in Pachyderm, however they
expose a version-control and history semantics similar to filesystem
objects, this is largely because, under the hood, they are implemented in
terms of filesystem objects. You can access previous versions of
a pipeline using the same ancestry syntax that works for commits and
branches, for example `pachctl inspect pipeline foo^` will give you the
previous version of the pipeline `foo`, `pachctl inspect pipeline foo.1`
will give you the first ever version of that same pipeline. This syntax
can be used wherever pipeline names are accepted. A common workflow is
reverting a pipeline to a previous version, this can be accomplished with:

```sh
$ pachctl extract pipeline pipeline^ | pachctl create pipeline
```

Historical versions of pipelines can also be returned with a `--history`
flag passed to `pachctl list pipeline` for example:

```sh
$ pachctl list pipeline --history -1
NAME      VERSION INPUT     CREATED     STATE / LAST JOB
Pipeline2 1       input2:/* 4 hours ago running / success
Pipeline1 3       input1:/* 4 hours ago running / success
Pipeline1 2       input1:/* 4 hours ago running / success
Pipeline1 1       input1:/* 4 hours ago running / success
```

## Jobs

Jobs do not have versioning semantics associated with them, however, they
are associated strongly with the pipelines that created them and thus
inherit some of their versioning semantics. This is reflected in `pachctl list
job`, by default this command will return all jobs from the most recent
versions of all pipelines. You can focus it on a single pipeline by passing `-p
pipeline` and you can focus it on a previous version of that pipeline by
passing `-p pipeline^`. Furthermore you can get jobs from multiple versions of
pipelines by passing the `--history` flag, for example: `pachctl list job
--history -1` will get you all jobs from all versions of all pipelines.
