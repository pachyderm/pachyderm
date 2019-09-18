# Commit

A commit is a snapshot that preserves the state of your data at a point in time.
It represents a single set of changes to files or directories
in your Pachyderm repository. A commit is a user-defined operation, which means
that you can start a commit, make changes, and then close the commit
after you are done.

Each commit has a unique identifier (ID) that you can reference in
the `<repo>@<commitID or branch>` format. When you create a new
commit, the previous commit on which the new commit is based becomes
the parent of the new commit.

You can obtain information about commits in a repository by running
`pachctl list commit <repo>` or `pachctl inspect commit <commitID>`.
In Pachyderm, commits are atomic operations that capture a state of
the files and directories in a repository. Unlike Git commits, Pachyderm
commits are centralized and transactional. You can start a commit by running
the `pachctl start commit` command, make changes to the repository, and close
the commit by running the `pachctl finish commit` command. After the commit is
finished, Pachyderm saves the new state of the repository.

When you *start*, or *open*, a commit, it means that you can make changes
by using `put file`, `delete file`, or other commands. You can
*finish*, or *close* a commit, which means the commit is immutable and
cannot be changed.

The `pachctl list commit repo@branch` command. This command returns a
timestamp, size, parent, and other information about the commit.
The initial commit has `<none>` as a parent.

**Example:**

```bash
pachctl list commit images@master
REPO   BRANCH COMMIT                           PARENT                           STARTED        DURATION           SIZE
raw_data master 8248d97632874103823c7603fb8c851c 22cdb5ae05cb40868566586140ea5ed5 6 seconds ago  Less than a second 5.121MiB
raw_data master 22cdb5ae05cb40868566586140ea5ed5 <none>                           33 minutes ago Less than a second 2.561MiB
```

The `list commit <repo>` command displays all commits in all branches
in the specified repository.

The `pachctl inspect commit` command enables you to view detailed
information about a commit, such as the size, parent, and the original
branch of the commit, as well as how long ago the commit was
started and finished. The `--full-timestamps` flag, enables you to
see the exact date and time of when the commit was opened and when it
was finished.
If you specify a branch instead of a specific commit, Pachyderm
displays the information about the HEAD of the branch.
For most commands, you can specify either a branch or a commit ID.

The most important information that the `pachctl inspect commit`
command provides the origin of the commit, or its provenance.
Typically, provenance can be tracked for the commits in the
output repositories.

**Example:**

```bash
$ pachctl inspect commit edges@master
Commit: edges@234fd5ea60ae4422b075f1945786ec5f
Original Branch: master
Parent: c5c9849ebb5849dc8bf37c4a925e3b20
Started: 12 seconds ago
Finished: 7 seconds ago
Size: 22.22KiB
Provenance:  images@e55ab0f1c44544ecb93151e11867c6b5 (master)  __spec__@ede9d05de2584108b03593301f1fdf81 (edges)
```

The `delete commit` command enables you to delete opened and closed
commits, which results in permanent loss of all the data *introduced* in
those commits. The `delete commit` command makes it as the deleted
commit never happened. If the deleted commit was the HEAD of the
branch, its parent becomes the HEAD.
You can only delete a commit from an input repository at the top of
your commit history, also known as DAG. Deleting a commit in the middle
of a DAG breaks the provenance chain. When you delete a commit from
the top of your DAG, Pachyderm automatically deletes all the commits
that were created in downstream output repos by processing the deleted commit,
making it as if the commit never existed.

Commit deletion is an irreversible operation that should be used with caution.
An alternative and a much safer way to revert incorrect data changes is to
move the HEAD of the branch or create a new commit that removes
the incorrect data.

**Example:**

```bash
$ pachctl delete commit raw_data@8248d97632874103823c7603fb8c851c
```

**See also:**

- [Provenance](provenance.md)
