# Commit

A commit is a snapshot that preserves the state of your data at a point in time.
It represents a single set of changes to files or directories
in your Pachyderm repository. Commit is a user-defined operation, which means
that you can start a commit, make changes, and then close the commit
after you are done.

Each commit has a unique identifier (ID) that you can reference in
the `<repo>@<commitID>` format. When you create a new
commit, the previous commit on which the new commit is based becomes
the parent of the new commit. Your pipeline history
consists of those parent-child relationships between your data commits.

You can obtain information about commits in a repository by running
`list commit <repo>` or `inspect commit <commitID>`.
In Pachyderm, commits are atomic operations that capture a state of
the files and directories in a repository. Unlike Git commits, Pachyderm
commits are centralized and transactional. You can start a commit by running
the `pachctl start commit` command, make changes to the repository, and close
the commit by running the `pachctl finish commit` command. After the commit is
finished, Pachyderm saves the new state of the repository.

When you *start*, or *open*, a commit, it means that you can make changes
by using `put file`, `delete file`, or other commands. You can
*finish*, or *close* a commit which means the commit is immutable and
cannot be changed.

The `pachctl list commit repo@branch` command returns a
timestamp, size, parent, and other information about the commit.
The initial commit has `<none>` as a parent.

!!! example
    ```shell
    pachctl list commit images@master
    ```

    **System Response:**

    ```shell
    REPO     BRANCH COMMIT                           PARENT                           STARTED        DURATION           SIZE
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

!!! example
    ```shell
    pachctl inspect commit raw_data@master --full-timestamps
    ```

    **System Response:**

    ```shell
    Commit: raw_data@8248d97632874103823c7603fb8c851c
    Original Branch: master
    Parent: 22cdb5ae05cb40868566586140ea5ed5
    Started: 2019-07-29T18:09:51.397535516Z
    Finished: 2019-07-29T18:09:51.500669562Z
    Size: 5.121MiB
    ```

The `delete commit` command enables you to delete opened and closed
commits, which results in permanent loss of all the data introduced in
those commits. You can think about the `delete commit` command as an
equivalent of the `rm -rf` command in Linux.
It is an irreversible operation that should be used with caution.
An alternative and a much safer way to revert incorrect data changes is to
move the `HEAD` of the branch or create a new commit that removes
the incorrect data.

!!! example
    ```shell
    pachctl delete commit raw_data@8248d97632874103823c7603fb8c851c
    ```
