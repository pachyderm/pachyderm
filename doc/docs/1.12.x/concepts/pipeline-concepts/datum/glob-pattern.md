# Glob Pattern

Defining how your data is spread among workers is one of
the most important aspects of distributed computation.

Pachyderm uses **glob patterns** to provide flexibility to
define data distribution.

You can think of each input repository as a filesystem where
the glob pattern is applied to the root. 
The files and directories that match the
glob pattern constitute the [datums](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/)
that will be processed by the worker(s) that run your pipeline code.

!!! Important
        You must **configure a glob pattern for each PFS input** of a [pipeline specification](). 

!!! Note
     The Pachyderm's concept of glob patterns is similar to the Unix glob patterns.
     For example, the `ls *.md` command matches all files with the
     `.md` file extension.


In Pachyderm, the `/` and `*` indicators are most
commonly used globs.

Let's list the glob patterns at your disposal. We will later illustrate their use in an example:

| Glob Pattern     | Datum created|
|-----------------|---------------------------------|
| `/` | Pachyderm denotes the **whole repository as a single datum** and sends all input data to a single worker node to be processed together.|
| `/*`| Pachyderm defines **each top-level filesystem object**, a file or a directory in the input repo, **as a separate datum**. For example, if you have a repository with ten files and no directory structure, Pachyderm identifies each file as a single datum and processes them independently.|
| `/*/*`| Pachyderm processes **each filesystem object in each subdirectory as a separate datum**.|
| `/**` | Pachyderm processes **each filesystem object in all directories and subdirectories as a separate datum**.|

Glob patterns also let you take only a particular directory or subset of
directories as an input instead of the whole repo.
We will elaborate on this more in the following example.

If you have more than one input repo in your pipeline,
you can define a different glob pattern for each input
repo. You can additionally combine the datums from each input repo
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
| `/`| This pattern matches `/`, the root directory itself, meaning **all the data would be one large datum**. All changes in any of the files and directories trigger Pachyderm to process the whole repository contents as a single datum.|*If you add a new file `Sacramento.json` to the `California/` directory, Pachyderm processes all changed files and folders in the repo as a single datum.*|
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
        - To understand how Pachyderm scales, read [Distributed Computing](https://docs.pachyderm.com/latest/concepts/advanced-concepts/distributed_computing/).
        - To learn about Datums' incremental processing, read our [Datum Processing](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/#datum-processing) section.
## Test a Glob pattern

You can use the `pachctl glob file` command to preview which filesystem
objects a pipeline defines as datums. This command helps
you to test various glob patterns before you use them in a pipeline.

* If you set the `glob` property to `/`, Pachyderm detects all
top-level filesystem objects in the `train` repository as one
datum:

!!! Example
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

The number of datums defines how they are distributed across a job's available workers. 
Pachyderm allows you to check those datums:

  - before you create a pipeline 
  - for a past job 

### Testing a glob pattern before creating a pipeline
You can use the `pachctl list datum -f <my_pipeline_spec.json>` command to preview the datums created by a pipeline. 

!!! note "Note"  
    The pipeline does not need to have been created for the command to return the list of datums. This "dry run" helps you adjust your glob pattern when creating your pipeline specification file.
 

!!! example
    ```shell
    pachctl list datum -f edges.json
    ```
    **System Response:**

    ```shell
        ID FILES                                                STATUS TIME
    -  images@8c958d1523f3428a98ac97fbfc367bae:/g2QnNqa.jpg -      -
    -  images@8c958d1523f3428a98ac97fbfc367bae:/8MN9Kg0.jpg -      -
    -  images@8c958d1523f3428a98ac97fbfc367bae:/46Q8nDz.jpg -      -
    ```

### Running list datum on a past job 
You can use the `pachctl list datum <job_number>` command to check the datums processed by a given job.

!!! example
    ```shell
    pachctl list datum d10979d9f9894610bb287fa5e0d734b5
    ```
    **System Response:**

    ```shell
        ID                                                                   FILES                                                STATUS TIME
    ebd35bb33c5f772f02d7dfc4735ad1dde8cc923474a1ee28a19b16b2990d29592e30 images@8c958d1523f3428a98ac97fbfc367bae:/g2QnNqa.jpg -      -
    ebd3ce3cdbab9b78cc58f40aa2019a5a6bce82d1f70441bd5d41a625b7769cce9bc4 images@8c958d1523f3428a98ac97fbfc367bae:/8MN9Kg0.jpg -      -
    ebd32cf84c73cfcc4237ac4afdfe6f27beee3cb039d38613421149122e1f9faff349 images@8c958d1523f3428a98ac97fbfc367bae:/46Q8nDz.jpg -      -
    ```

!!! note "Note"  
    Now that the 3 datums have been processed, their ID field is showing.

!!! info "Enabling Stats"
    - Running `list datum` on a given job execution of a pipeline that [enables stats](https://docs.pachyderm.com/latest/enterprise/stats/#enabling-stats-for-a-pipeline) allows you to additionally display the STATUS (running, failed, success) and TIME of each datum.
    - You might want to follow up with [inspect datum <ID>](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_inspect_datum/) to detail the files that a specific datum includes.