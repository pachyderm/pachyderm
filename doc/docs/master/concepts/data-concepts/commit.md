# Commit

!!! Note "Attention"
         Note that Pachyderm uses the term `commit` at two different levels. A global level (check [GlobalID](../globalID) for more details) and commits that occur in a repository. The following page details the latter. 
## Definition

In Pachyderm, commits are atomic operations that **snapshot and preserve the state of
the files and directories in a repository** at a point in time. 
Unlike Git commits, Pachyderm commits are centralized and transactional.
You can start a commit by running the `pachctl start commit` command with reference
to a specific repository. 
After you're done making changes to the repository (`put file`, `delete file`, ...),
you can finish your modifications by running the `pachctl finish commit` command.
This command saves your changes and closes that repository's commit,
indicating the data is ready for processing by downstream pipelines.

!!! Warning
    `start commit` can only be used on input repos without [provenance](./provenance.md). Such repos are the entry points of a DAG.
    You cannot manually start a commit from a pipeline [output or meta repo](./repo.md).

 When you create a new commit, the previous commit on which the new commit is based becomes the parent of the new commit. Your repo history consists of those parent-child relationships between your data commits.

!!! Note
    An initial commit has `<none>` as a parent.

Additionally, **commits have an "origin"**.
You can see an origin as the answer to: **"What triggered the production of this commit?"**.

That origin can be of 3 types:

- `USER`: The commit is the result of a user change (`put file`, `update pipeline`, `delete file`...)
!!! Info
    Every initial change is a `USER` change.
- `AUTO`: Pachyderm's pipelines are data-driven. A data commit to a data repository may
    trigger downstream processing jobs in your pipeline(s). The output commits from
    triggered jobs will be of type `AUTO`.
- `ALIAS`: Neither `USER` nor `AUTO` - `ALIAS` commits are essentially placeholder commits.
    They have the same content as their parent commit and are mainly used for [global IDs](../globalID/).


!!! Warning "Important Note"
    To track provenance, Pachyderm requires **all commits to belong to exactly one branch**.
    When moving a commit from one [branch](./branch.md) to another, Pachyderm creates an `ALIAS` commit on the other branch.


Each commit has an alphanumeric identifier (ID) that you can reference in the `<repo>@<commitID>` format (or `<repo>@<branch>=<commitID>` if the commit has multiple branches from the same repo) .

You can obtain information about all commits in Pachyderm
by running `list commit` or `inspect commit <commitID>`.
To restrict to a particular repository, use `list commit <repo>`,
`list commit <repo>@<branch>`, or `inspect commit <repo>@<commitID>`.

## List commits
The `pachctl list commit` command returns list of all commits, along with the number of
sub-commits they contain, when they were created, and when a sub-commit was most recently modified.

!!! example
    ```shell
    pachctl list commit
    ```

    **System Response:**

    ```shell
    ID                               COMMITS PROGRESS CREATED        MODIFIED
    c6d7be4a13614f2baec2cb52d14310d0 7       ▇▇▇▇▇▇▇▇ 13 seconds ago 13 seconds ago
    385b70f90c3247e69e4bdadff12e44b2 7       ▇▇▇▇▇▇▇▇ 2 hours ago    13 seconds ago
    ```

The `pachctl list commit <repo>@<branch>` command returns the sub-commits in the given
branch of a repo, including all commits.

!!! example
    ```shell
    pachctl list commit images@master
    ```

    **System Response:**

    ```shell
    REPO   BRANCH COMMIT                           FINISHED        SIZE       ORIGIN DESCRIPTION
    images master c6d7be4a13614f2baec2cb52d14310d0 33 minutes ago  5.121MiB    USER
    images master 385b70f90c3247e69e4bdadff12e44b2 2 hours ago     2.561MiB    USER
    ```

`list commit <repo>`, without mention of a branch, displays results from all branches of the specified repository.

## Inspect commit
The `pachctl inspect commit <commitID>` command enables you to view a list of sub-commits in a commit.

!!! example
    ```shell
    $ pachctl inspect commit images@c6d7be4a13614f2baec2cb52d14310d0
    ```

    **System Response:**

    ```shell
    REPO         BRANCH COMMIT                           FINISHED       SIZE  ORIGIN DESCRIPTION
    edges.spec   master c6d7be4a13614f2baec2cb52d14310d0 25 seconds ago <= 0B ALIAS
    montage.spec master c6d7be4a13614f2baec2cb52d14310d0 25 seconds ago <= 0B ALIAS
    images       master c6d7be4a13614f2baec2cb52d14310d0 25 seconds ago <= 0B USER
    edges        master c6d7be4a13614f2baec2cb52d14310d0 16 seconds ago <= 0B AUTO
    edges.meta   master c6d7be4a13614f2baec2cb52d14310d0 16 seconds ago <= 0B AUTO
    montage      master c6d7be4a13614f2baec2cb52d14310d0 13 seconds ago <= 0B AUTO
    montage.meta master c6d7be4a13614f2baec2cb52d14310d0 13 seconds ago <= 0B AUTO
    ```

The `pachctl inspect commit repo@commitID` command enables you to view detailed
information about a sub-commit (size, parent, the branch it belongs to,
how long ago the commit was started and finished...).

- The `--full-timestamps` flag will give you the exact date and time
of when the sub-commit was opened and finished.
- If you specify a branch instead of a specific commit (`pachctl inspect commit repo@branch`),
Pachyderm displays the information about the HEAD of the branch.

!!! example
    Add a `--raw` flag to output a more detailed JSON version of the sub-commit.
    ```shell
    $ pachctl inspect commit images@c6d7be4a13614f2baec2cb52d14310d0 --raw
    ```

    **System Response:**

    ```json
    {
        "commit": {
            "branch": {
            "repo": {
                "name": "images",
                "type": "user"
            },
            "name": "master"
            },
            "id": "c6d7be4a13614f2baec2cb52d14310d0"
        },
        "origin": {
            "kind": "USER"
        },
        "parent_commit": {
            "branch": {
            "repo": {
                "name": "images",
                "type": "user"
            },
            "name": "master"
            },
            "id": "385b70f90c3247e69e4bdadff12e44b2"
        },
        "started": "2021-07-06T01:17:48.488831754Z",
        "finished": "2021-07-06T01:17:48.488831754Z",
        "details": {
            "size_bytes": "244068"
        }
    }
    ```

## Squash commit

See [`squash commit`](./globalID/#squash-commit).



