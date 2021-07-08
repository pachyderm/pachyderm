# Commit

## Definition

In Pachyderm, commits are atomic operations that **snapshot and preserve the state of
the files and directories in a repository** at a point in time. 
Unlike Git commits, Pachyderm commits are centralized and transactional. 
You can start a commit by running the `pachctl start commit` command, 
make changes to the repository (`put file`, `delete file`, ...), 
and close the commit by running the `pachctl finish commit` command. After the commit is
finished, Pachyderm saves the new state of the repository.

!!! Warning 
    `start commit` can only be used on input repos (i.e. repos without [provenance](./provenance.md)).
    You cannot manually start a commit in a pipeline [output or meta repo](./repo.md).


When you create a new commit, the **previous commit on which the new commit is based becomes
the parent** of the new commit. Your repo history
consists of those parent-child relationships between your data commits.

!!! Note
    An initial commit has `<none>` as a parent.

Additionally, **commits have an "origin"**. 
You can see an origin as the answer to: **"What triggered the production of this commit"**. 

That origin can be of 3 types:

- `USER`: The commit is the result of a user change (`put file`, `update pipeline`, `delete file`...)
!!! Info
    Every initial change is a `USER` change.
- `AUTO`: Pachyderm's pipelines being data-driven, the initial commit will automatically trigger downstream processing jobs in your pipeline. Those output commits will be of `AUTO` origin.
- `ALIAS`: Neither `USER` nor `AUTO` - `ALIAS` commits are essentially empty commits. They have the same content than their parent commit and are mainly used for [global IDs](). 


!!! Warning "Important Note"
    **All commits must exist on exactly one branch**.  
    When moving a commit from one [branch](./branch.md) to another, Pachyderm creates an alias commit on the other branch.


Each commit has an identifier (ID) that you can reference in
the `<repo>@<commitID>` format.

You can obtain information about commits in a repository by running
`list commit <repo>` or `inspect commit <commitID>`.

## List commits
The `pachctl list commit repo@branch` command returns the
list of commits in the given branch of a repo.


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

    `list commit <repo>`, without mention of a branch, displays all commits in all branches of the specified repository.

## Inspect commit
The `pachctl inspect commit repo@commitID` command enables you to view detailed
information about a commit (size, parent, the original
branch of the commit, how long ago the commit was
started and finished...). 

- The `--full-timestamps` flag will give you the exact date and time
of when the commit was opened and finished.
- If you specify a branch instead of a specific commit (`pachctl inspect commit repo@branch`), 
Pachyderm displays the information about the HEAD of the branch.

!!! example
    Add a `--raw` flag to output a more detailed JSON version of the commit.
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

See [`squash commitset`]().



