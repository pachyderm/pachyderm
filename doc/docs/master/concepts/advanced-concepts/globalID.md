# Global Identifier

## Definition
Pachyderm provides users with a simple way to follow a change throughout their DAG (i.e., Traverse Provenance and Subvenance).

When you commit data to Pachyderm, your new commit has an ID associated with it that you can easily check by running `pachctl list commit repo@branch`. 
**All resulting downstream commits and jobs will then share that ID (Global Identifier).**

!!! Info "TLDR"
    Commit and job IDs **represent a logically-related set of objects**. 
    The ID of a commit is also the ID of any commits created along with it due to provenance relationships, 
    as well as the ID of any jobs that were triggered because of the creation of those commits. 

This ability to track down related commits and jobs with one global identifier brought the need to introduce a new scope to our original concepts of [job](./job.md) and [commit](./commit.md):

- A commit with a `global` scope (**global commit**), referred to in our CLI as `pachctl list commit <commitID>` represents "the set of all provenance-dependent commits sharing the same ID". The same term of `commit`, applied to the more focused scope of a repo (`pachctl list commit <repo>@<commitID>` or `pachctl list commit <repo>@, branch>=<commitID>`), represents "a Git-like record of the state of a single repository's file system".
Similarly, the same nuances in the scope of a job give the term two possible meanings:
- A job with a `global` scope (**global job**),  referred to in our CLI as `pachctl list job <commitID>`, is "the set of jobs created due to commits in a global commit". Narrowing down the scope to a single pipeline (`pachctl list job <pipeline>@<commitID>`) shifts the meaning to "an execution of a given pipeline of your DAG".

Using this global identifier you can:

## List All Global Commits And Global Jobs
You can list all global commits by running the following command:
```shell
$ pachctl list commit
```
Each global commit displays how many (sub) commits it is made of.
```
ID                               SUBCOMMITS PROGRESS CREATED        MODIFIED
1035715e796f45caae7a1d3ffd1f93ca 7          ▇▇▇▇▇▇▇▇ 7 seconds ago  7 seconds ago
28363be08a8f4786b6dd0d3b142edd56 6          ▇▇▇▇▇▇▇▇ 24 seconds ago 24 seconds ago
e050771b5c6f4082aed48a059e1ac203 4          ▇▇▇▇▇▇▇▇ 24 seconds ago 24 seconds ago
```
Similarly, if you run the equivalent command for global jobs:
```shell
$ pachctl list job
```
you will notice that the job IDs are shared with the global commit IDs.

```
ID                               SUBJOBS PROGRESS CREATED            MODIFIED
1035715e796f45caae7a1d3ffd1f93ca 2       ▇▇▇▇▇▇▇▇ 55 seconds ago     55 seconds ago
28363be08a8f4786b6dd0d3b142edd56 1       ▇▇▇▇▇▇▇▇ About a minute ago About a minute ago
e050771b5c6f4082aed48a059e1ac203 1       ▇▇▇▇▇▇▇▇ About a minute ago About a minute ago
```
For example, in this example, 7 commits and 2 jobs are involved in the changes occured
in the global commit ID 1035715e796f45caae7a1d3ffd1f93ca.

!!! Note
        The progress bar is equally divided to the number of steps, or pipelines,
        you have in your DAG. In the example above, `1035715e796f45caae7a1d3ffd1f93ca` is two steps.
        If one of the sub-jobs fails, you will see the progress bar turn red
        for that pipeline step. To troubleshoot, look into that particular
        pipeline execution.

## List All Commits And Jobs With A Global ID

To list all (sub) commits involved in a global commit:
```shell
$ pachctl list commit 1035715e796f45caae7a1d3ffd1f93ca
```
```
REPO         BRANCH COMMIT                           FINISHED      SIZE        ORIGIN DESCRIPTION
images       master 1035715e796f45caae7a1d3ffd1f93ca 5 minutes ago 238.3KiB    USER
edges.spec   master 1035715e796f45caae7a1d3ffd1f93ca 5 minutes ago 244B        ALIAS
montage.spec master 1035715e796f45caae7a1d3ffd1f93ca 5 minutes ago 405B        ALIAS
montage.meta master 1035715e796f45caae7a1d3ffd1f93ca 4 minutes ago 1.656MiB    AUTO
edges        master 1035715e796f45caae7a1d3ffd1f93ca 5 minutes ago 133.6KiB    AUTO
edges.meta   master 1035715e796f45caae7a1d3ffd1f93ca 5 minutes ago 373.9KiB    AUTO
montage      master 1035715e796f45caae7a1d3ffd1f93ca 4 minutes ago 1.292MiB    AUTO
```

Similarly, change `commit` in `job` to list all (sub) jobs linked to your global job ID.
```shell
$ pachctl list job 1035715e796f45caae7a1d3ffd1f93ca
```
```
ID                               PIPELINE STARTED       DURATION  RESTART PROGRESS  DL       UL       STATE
1035715e796f45caae7a1d3ffd1f93ca montage  5 minutes ago 4 seconds 0       1 + 0 / 1 79.49KiB 381.1KiB success
1035715e796f45caae7a1d3ffd1f93ca edges    5 minutes ago 2 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
```
For each pipeline execution (sub job) within this global job, Pachyderm shows the time since each sub job started and its duration, the number of datums in the **PROGRESS** section,  and other information.
The format of the progress column is `DATUMS PROCESSED + DATUMS SKIPPED / TOTAL DATUMS`.

For more information, see [Datum Processing States](../../../concepts/pipeline-concepts/datum/datum-processing-states/).

!!! Note
     The global commit and global job above have been created after
     a `pachctl put file images@master -i images.txt` in the images repo of the open cv example.

The following diagram illustrates the global commit and its various components:
    ![global_commit_after_putfile](../images/global_commit_after_putfile.png)

Let's explain the origin of each commit.

1. Inspect the commit ID 1035715e796f45caae7a1d3ffd1f93ca in the `images` repo,  the repo in which our change (`put file`) has originated:

    ```shell
    $ pachctl inspect commit images@1035715e796f45caae7a1d3ffd1f93ca --raw
    ```
    Note that this original commit is of `USER` origin (i.e., the result of a user change).

    !!! Note
        The list of all commit types is detailed in the [`Commit` page](../data-concepts/commit.md) of this section.

    ```json
    "origin": {
    "kind": "USER"
        },
    ```

1. Inspect the following commit 1035715e796f45caae7a1d3ffd1f93ca produced in the output repos of the edges pipeline:
    ```shell
    $ pachctl inspect commit edges@1035715e796f45caae7a1d3ffd1f93ca --raw
    ```
    ```json
    {
        "commit": {
            "branch": {
            "repo": {
                "name": "edges",
                "type": "user"
            },
            "name": "master"
            },
            "id": "1035715e796f45caae7a1d3ffd1f93ca"
        },
        "origin": {
            "kind": "AUTO"
        },
        "parent_commit": {
            "branch": {
            "repo": {
                "name": "edges",
                "type": "user"
            },
            "name": "master"
            },
            "id": "28363be08a8f4786b6dd0d3b142edd56"
        },
        "started": "2021-07-07T13:52:34.140584032Z",
        "finished": "2021-07-07T13:52:36.507625440Z",
        "direct_provenance": [
            {
            "repo": {
                "name": "edges",
                "type": "spec"
            },
            "name": "master"
            },
            {
            "repo": {
                "name": "images",
                "type": "user"
            },
            "name": "master"
            }
        ],
        "details": {
            "size_bytes": "22754"
        }
    }

    ```
    Note that the origin of the commit is of kind **`AUTO`** as it has been trigerred by the arrival of a commit in the upstream repo `images`.
    ```json
        "origin": {
            "kind": "AUTO"
        },
    ```

    The same origin (`AUTO` ) applies to the commits sharing that same ID in the `montage` output repo as well as `edges.meta` and `montage.meta` system repos. 
    !!! Note
        The list of all types of repos is detailed in the [`Repo` page](../data-concepts/repo.md) of this section.

- Besides  the `USER` and `AUTO` commits, notice a set of `ALIAS` commits in `edges.spec` and `montage.spec`:
```shell
    $ pachctl inspect commit edges.spec@336f02bdbbbb446e91ba27d2d2b516c6 --raw
```
The version of each pipeline within their respective `.spec` repos are neither the result of a user change, nor of an automatic change.
They have, however, contributed to the creation of the previous `AUTO` commits. 
To make sure that we have a complete view of all the data and pipeline versions involved in all the commits resulting from the initial 
`put file`, their version is kept as `ALIAS` commits under the same global ID.

For a fuller view of GlobalID in action, take a look at our [GlobalID illustration](https://github.com/pachyderm/pachyderm/tree/master/examples/globalID).

## Track Provenance Downstream

Pachyderm provides the `wait commit <commitID>` command that enables you
to track your commits downstream as they are produced. 

Unlike the `list commit <commitID>`, each line is printed as soon as a new (sub) commit of your global commit finishes.

Change `commit` in `job` to list the jobs related to your global job as they finish processing a commit.

## Squash A Global Commit

`pachctl squash commit 1035715e796f45caae7a1d3ffd1f93ca`
**combines all the file changes in the commits of a global commit
into their children** and then removes the global commit.
This behavior is inspired by the squash option in git rebase.
No data stored in PFS is removed.

