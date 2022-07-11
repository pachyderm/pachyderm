# Glob Pattern

Defining how your data is spread among workers is one of
the most important aspects of distributed computation.

Pachyderm uses **glob patterns** to provide flexibility to
define data distribution. 

!!! Note
     Pachyderm's concept of glob patterns is similar to Unix glob patterns.
     For example, the `ls *.md` command matches all files with the
     `.md` file extension.


The glob pattern applies to all of the directories/files in the branch specified by the [`pfs` section of the pipeline specification (referred to as PFS inputs)](../#pfs-input-and-glob-pattern). The directories/files that match are the [datums](../) that will be processed by the worker(s) that run your pipeline code. 

!!! Important
        You must **configure a glob pattern for each PFS input** of a [pipeline specification](../../../../reference/pipeline-spec/#pipeline-specification). 


We have listed some commonly used glob patterns. We will later illustrate their use in an example:

| Glob Pattern     | Datum created|
|-----------------|---------------------------------|
| `/` | Pachyderm denotes the **whole repository as a single datum** and sends all input data to a single worker node to be processed together.|
| `/*`| Pachyderm defines **each top-level files / directories** in the input repo, **as a separate datum**. For example, if you have a repository with ten files and no directory structure, Pachyderm identifies each file as a single datum and processes them independently.|
| `/*/*`| Pachyderm processes **each file / directory in each subdirectories as a separate datum**.|
| `/**` | Pachyderm processes **each file in all directories and subdirectories as a separate datum**.|

Glob patterns also let you take only a particular
subset of the files / directories, a specific branch...
as an input instead of the whole repo.
We will elaborate on this more in the following example.

If you have more than one input repo in your pipeline, 
or want to consider more than one branch in a repo, or any combination of the above,
you can define a different glob pattern for each PFS input. 
You can additionally combine the resulting datums
by using the `cross`, `union`, `join`, or `group` operator to
create the final datums that your code processes.
For more information, see [Cross and Union](./cross-union.md), [Join](./join.md), [Group](./group.md).

## Example of Defining Datums
Let's consider an input repo with the following structure where each top-level directory represents a US
state with a `json` file for each city in that state:

```
    /California
       /San-Francisco.json
       /Los-Angeles.json
       ...
    /Colorado
       /Denver.json
       /Boulder.json
       ...
    /Washington
        /Seattle.json
        /Vancouver.json
```

Now let's consider what the following glob patterns would match respectively:

|Glob Pattern| Corresponding match| Example|
|-----------------|---------------------------------||
| `/`| This pattern matches `/`, the root directory itself, meaning **all the data would be one large datum**. All changes in any of the files and directories trigger Pachyderm to process the whole repository contents as a single datum.|*If you add a new file `Sacramento.json` to the `California/` directory, Pachyderm processes all changed files and directories in the repo as a single datum.*|
| `/*`| This pattern matches **everything under the root directory**. It defines **one datum per state**, which means that all the cities for a given state are processed together by a single worker, but each state is processed independently.|*If you add a new file `Sacramento.json` to the `California/` directory, Pachyderm processes the `California/` datum only*.|
| `/Colorado/*`| This pattern matches **files only under the `/Colorado` directory**. It defines **one datum per city**.|*If you add a new file `Alamosa.json` to the `Colorado/` directory and `Sacramento.json` to the `California/` directory, Pachyderm processes the `Alamosa.json` datum only.*|
| `/C*`|  This pattern matches all **files under the root directory that start with the character `C`.**| *In the example, the `California` and  `Colorado` directories will each define a datum.*|
| `/*/*`|  This pattern matches **everything that's two levels deep relative to the root**.|*If we add County sub-directories to our states, `/California/LosAngeles/LosAngeles.json`, `/California/LosAngeles/Malibu.json` and `/California/SanDiego/LaMosa.json` for example, then this pattern would match each of those 3 .json files individually.*|
| `/**`| The match is applied at **all levels of your directory structure**. This is a recursive glob pattern. Let's look at the additional example below for more detail.||


!!! example "Example: The case of the `/**` glob pattern"
    

    Say we have the following repo structure:
    ```
      /nope1.txt
      /test1.txt
      /foo-1
        /nope2.txt
        /test2.txt
      /foo-2
        /foo-2_1
          /nope3.txt
          /test3.txt
          /anothertest.txt
    ```
    ...and apply the following pattern to our input repo:
    ```
      "glob": "/**test*.txt"
    ```
    We are **recursively matching all `.txt` files containing `test`** starting from our input repo's root directory.
    In this case, the resulting datums will be:
    
    ```
      - /test1.txt
      - /foo-1/test2.txt
      - /foo-2/foo-2_1/test3.txt
      - /foo-2/foo-2_1/anothertest.txt
    ```

!!! See "See Also"
        - To understand how Pachyderm scales, read [Distributed Computing](https://docs.pachyderm.com/latest/concepts/advanced-concepts/distributed-computing/){target=_blank}.
        - To learn about Datums' incremental processing, read our [Datum Processing](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/#datum-processing){target=_blank} section.

## Test a Glob pattern

You can use the `pachctl glob file` command to preview which filesystem
objects a pipeline defines as datums. This command helps
you to test various glob patterns before you use them in a pipeline.

* If you set the `glob` property to `/`, Pachyderm detects all
top-level filesystem objects in the `train` repository as one
datum:

!!! example
    ```shell
    pachctl glob file train@master:/
    ```

    **System Response:**

    ```shell
    NAME TYPE SIZE
    /    dir  15.11KiB
    ```

* If you set the `glob` property to `/*`, Pachyderm detects each
top-level filesystem object in the `train` repository as a separate
datum:

!!! example
    ```shell
    pachctl glob file train@master:/*
    ```

    **System Response:**

    ```shell
    NAME                   TYPE SIZE
    /IssueSummarization.py file 1.224KiB
    /requirements.txt      file 74B
    /seq2seq_utils.py      file 13.81KiB
    ```

## Test your Datums

The granularity of your datums defines how your data will be distributed across the available workers allocated to a job.
Pachyderm allows you to check those datums:

  - for a pipeline currently being developed  
  - for a past job 

### Testing your glob pattern before creating a pipeline
You can use the `pachctl list datum -f <my_pipeline_spec.json>` command to preview the datums defined by a pipeline given its specification file. 

!!! note "Note"  
    The pipeline does not need to have been created for the command to return the list of datums. This "dry run" helps you adjust your glob pattern when creating your pipeline.
 

!!! example
    ```shell
    pachctl list datum -f edges.json
    ```
    **System Response:**

    ```
      ID FILES                                                STATUS TIME
    -  images@b8687e9720f04b7ab53ae8c64541003b:/46Q8nDz.jpg -      -
    -  images@b8687e9720f04b7ab53ae8c64541003b:/8MN9Kg0.jpg -      -
    -  images@b8687e9720f04b7ab53ae8c64541003b:/Togu2RY.jpg -      -
    -  images@b8687e9720f04b7ab53ae8c64541003b:/g2QnNqa.jpg -      -
    -  images@b8687e9720f04b7ab53ae8c64541003b:/w7RVTsv.jpg -      -
    ```

### Running list datum on a past job 
You can use the `pachctl list datum <pipeline>@<job_ID>` command to check the datums processed by a given job.

!!! example
    ```shell
    pachctl list datum edges@b8687e9720f04b7ab53ae8c64541003b
    ```
    **System Response:**

    ```
      ID                                                               FILES                                                STATUS  TIME
    a4149cd1907145f982e0eb49c50af3f1d4d8fecaa8647d62f2d9d93e30578df8 images@b8687e9720f04b7ab53ae8c64541003b:/w7RVTsv.jpg success Less than a second
    e2b4628dd88b179051ba0576e06fac12ae2e4d16165296212d0e98de501d17df images@b8687e9720f04b7ab53ae8c64541003b:/Togu2RY.jpg success Less than a second
    353b6d2a5ac78f56facc7979e190affbb8f75c6f74da84b758216a8df77db473 images@7856142de8714c11b004610ea7af2378:/8MN9Kg0.jpg skipped Less than a second
    b751702850acad5502dc51c3e7e7a1ac10ba2199fdb839989cd0c5430ee10b84 images@fc9a12ee149a4499a1a7da0a31971b37:/46Q8nDz.jpg skipped Less than a second
    de9e3703322eff2ab90e89ff01a18c448af9870f17e78438c5b0f56588af9c44 images@7856142de8714c11b004610ea7af2378:/g2QnNqa.jpg skipped Less than a second
    ```

!!! note "Note"  
    In this example, you can see that the job `b8687e9720f04b7ab53ae8c64541003b` only processed 2 datums from the images input repo. The rest was skipped as it had been processed by previous jobs already. Notice that the ID of the datums is now showing.

!!! info "Stats and Datum Metadata"
    - Running `list datum` on a given job execution of a pipeline allows you to additionally display the STATUS (running, failed, success) and TIME of each datum.
    - You might want to follow up with [inspect datum pipeline@job_number datum ID](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_inspect_datum/){target=_blank} to detail the files that a specific datum includes.
    
        ```shell
        pachctl inspect datum edges@b8687e9720f04b7ab53ae8c64541003b a4149cd1907145f982e0eb49c50af3f1d4d8fecaa8647d62f2d9d93e30578df8
        ```
        **System Response:**
        ```
        ID	a4149cd1907145f982e0eb49c50af3f1d4d8fecaa8647d62f2d9d93e30578df8
        Job ID	b8687e9720f04b7ab53ae8c64541003b
        State	SUCCESS
        Data Downloaded	606.3KiB
        Data Uploaded	26.96KiB
        Total Time	582.000134ms
        Download Time	40.062075ms
        Process Time	535.387088ms
        Upload Time	6.550971ms
        PFS State:
          REPO         COMMIT                             PATH
          edges.meta   b8687e9720f04b7ab53ae8c64541003b   /pfs/a4149cd1907145f982e0eb49c50af3f1d4d8fecaa8647d62f2d9d93e30578df8
        Inputs:
          REPO     COMMIT                             PATH
          images   b8687e9720f04b7ab53ae8c64541003b   /w7RVTsv.jpg
        ```
        Add `--raw` for a full JSON version of the [datum's metadata](../metadata/).

