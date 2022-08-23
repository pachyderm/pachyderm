# Commit

!!! Note "Attention"
         Note that Pachyderm uses the term `commit` at two different levels. A global level (check [GlobalID](../../advanced-concepts/globalID){target=_blank} for more details) and commits that occur on the given branch of a repository. The following page details the latter. 

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
        Every `USER` change is an initial commit.

- `AUTO`: Pachyderm's pipelines are data-driven. A data commit to a data repository may
    trigger downstream processing jobs in your pipeline(s). The output commits from
    triggered jobs will be of type `AUTO`.
- `ALIAS`: Neither `USER` nor `AUTO` - `ALIAS` commits are essentially placeholder commits.
    They have the same content as their parent commit and are mainly used for [global IDs](../../advanced-concepts/globalID/).


!!! Note
    To track provenance, Pachyderm requires **all commits to belong to exactly one branch**.
    When moving a commit from one [branch](./branch.md) to another, Pachyderm creates an `ALIAS` commit on the other branch.


Each commit has an alphanumeric identifier (ID) that you can reference in the `<repo>@<commitID>` format (or `<repo>@<branch>=<commitID>` if the commit has multiple branches from the same repo) .

You can obtain information about all commits with a given ID
by running `pachctl list commit <commitID>` or restrict to a particular repository `pachctl list commit <repo>`,
`pachctl list commit <repo>@<branch>`, or `pachctl inspect commit <repo>@<commitID> --raw`.

## List Commits
- The `pachctl list commit` command returns list of all global commits. This command is detailed in [this section of Global ID](../../advanced-concepts/globalID/#list-all-global-commits-and-global-jobs).

- The `pachctl list commit <commitID>` commands returns the list of all commits sharing the same `<commitID>`. This command is detailed in [this section of Global ID](../../advanced-concepts/globalID/#list-all-commits-and-jobs-with-a-global-id). 

- Note that you can also track your commits downstream as they complete by running `pachctl wait commit <commitID>`. 

- The `pachctl list commit <repo>@<branch>` command returns the commits in the given branch of a repo.

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

- `list commit <repo>`, without mention of a branch, displays results from all branches of the specified repository.

## Inspect Commit
The `pachctl inspect commit <repo>@<commitID>` command enables you to view detailed
information about a commit in a given repo (size, parent, the branch it belongs to,
how long ago the commit was started and finished...).

- The `--full-timestamps` flag will give you the exact date and time
of when the commit was opened and finished.
- If you specify a branch instead of a specific commit (`pachctl inspect commit <repo>@<branch>`),
Pachyderm displays the information about the HEAD of the branch.

!!! example
    Add a `--raw` flag to output a detailed JSON version of the commit.
    ```shell
    pachctl inspect commit images@c6d7be4a13614f2baec2cb52d14310d0 --raw
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
        "started": "2021-08-02T20:13:10.393036120Z",
        "finishing": "2021-08-02T20:13:10.393036120Z",
        "finished": "2021-08-02T20:13:11.851931210Z",
        "size_bytes_upper_bound": "244068",
        "details": {
            "size_bytes": "244068"
        }
    }
    ```

## Squash And Delete Commit

See [`squash commit`](../../../how-tos/basic-data-operations/removing-data-from-pachyderm/#squash-non-head-commits) and  [`delete commit`](../../../how-tos/basic-data-operations/removing-data-from-pachyderm/#delete-the-head-of-a-branch) in the `Delete a Commit / Delete Data` page of the How-Tos section of this Documentation.





