>![pach_logo](../img/pach_logo.svg) INFO - Pachyderm 2.0 comes with significant architectural refactoring and functional enhancements. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Master Branch: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - 1.13.x Branch : Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

# Inner and Outer Join Inputs
>![pach_logo](./img/pach_logo.svg) The *outer join* input is available in version **1.12 and higher**.

- In our first example, we will create a pipeline whose input datums result from a simple `inner join` between 2 repos.
- In our second example, we will showcase three variations of `outer join` pipelines between 2 repos and outline how they differ from inner join and each other.

At the end of this page, you will understand the fundamental difference between the datums produced by an inner join and those created by an outer join.

***Table of Contents***

- [1. Getting ready](#1-getting-ready)
- [2. Data structure and naming convention](#2-data-structure-and-naming-convention)
- [3. Data preparation](#3-data-preparation)
- [4. Example 1 : An Inner Join pipeline creation](#4-example-1--an-inner-join-pipeline-creation)
- [5. Example 2 : Outer Join pipeline creation](#5-example-2--outer-join-pipeline-creation) 
    - [***Case 1*** Outer join on the Returns repo only](#case-1-outer-join-on-the-returns-repo-only)
    - [***Case 2*** Outer join on the Stores repo only](#case-2-outer-join-on-the-stores-repo-only)
    - [***Case 3*** Outer join on both the Stores and Returns repos](#case-3-outer-join-on-both-the-stores-and-returns-repos)

***Key concepts***

For these examples, we recommend being familiar with the following concepts:

- [Join](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/join/) pipelines - execute your code on files that match a specific naming pattern in your joined repos.
- [Glob patterns](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/glob-pattern/) - A "RegEx-like" string matching on file paths and names.

Additionally, make sure that you understand the concept of [datum](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/). 

## 1. Getting ready
***Prerequisite***
- A workspace on [Pachyderm Hub](https://docs.pachyderm.com/latest/pachhub/pachhub_getting_started/) (recommended) or Pachyderm running [locally](https://docs.pachyderm.com/latest/getting_started/local_installation/).
- [pachctl command-line ](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl) installed, and your context created (i.e. you are logged in)

***Getting started***
- Clone this repo.
- Make sure Pachyderm is running. You should be able to connect to your Pachyderm cluster via the `pachctl` CLI. 
Run a quick:
```shell
$ pachctl version

COMPONENT           VERSION
pachctl             2.0.1
pachd               2.0.1
```
Ideally, have your pachctl and pachd versions match. At a minimum, you should always use the identical major & minor versions of pachctl and pachd. 
## 2. Data structure And Naming Convention

>![pach_logo](./img/pach_logo.svg)  Remember, in Pachyderm, the join operates at the file-path level, **not** the files' content. Therefore, the structure of your directories and file naming conventions are key elements when implementing your use cases in Pachyderm.

We have derived our examples from simplified retail use cases: 
- Purchases and Returns are made in given Stores. 
- Those Stores have a given location (here, a zip code). 
- There are 0 to many Stores in a given zip code.

Let's take a look at our data structure and naming convention. 
We will create 3 repos:
* Repo: `stores` - Each store data are JSON files named after the storeID.

    ```shell
        └── STOREID1.txt
        └── STOREID2.txt
        └── STOREID3.txt
        └──  ...
    ```
    To further these examples beyond the creation of datum, we have added some simple processing at the code level. We will use the `zipcode` property (see JSON content of a Store file below) to aggregate our data. 

    This is what the content of one of those STOREIDx.txt files looks like.
    ```json
       {
           "storeid":"4",
            "name":"mariposa st.",
           "address":{
               "zipcode":"94107",
               "country":"US"
           }
       }
    ```

* Repo: `purchases` - Each purchase info is kept in a file named by concatenating the purchase's order number and its store ID.

    ```shell
        └── ORDERW080520_STOREID1.txt
        └── ORDERW080521_STOREID1.txt
        └── ORDERW078929_STOREID2.txt
        └── ...
    ```
* Repo: `returns` - Same naming convention as purchases.

    ```shell
        └── ORDERW080528_STOREID5.txt
        └── ODERW080520_STOREID1.txt
        └── ...
    ```
## 3. Data preparation
Let's start by creating our mock data and populating our repositories.
This preparatory step will allow you to experiment with the inner and outer join examples in any order.

***Step 1*** - Prepare your data

The setup target of the `Makefile` in `pachyderm/examples/joins` will create 3 directories (stores, purchases, and returns) containing our example's data.
In the `examples/joins` directory, run:
```shell
make setup
```
***Step 2*** - Create and populate Pachyderm's repositories from the directories above.

In the `examples/joins` directory, run:
```shell
make create
```
or run:
```shell
pachctl create repo stores
pachctl create repo purchases
pachctl create repo returns
pachctl put file -r stores@master:/ -f stores
pachctl put file -r purchases@master:/ -f purchases
pachctl put file -r returns@master:/ -f returns
```

Have a look at the content of your repositories: 
```shell
pachctl list file stores@master
pachctl list file purchases@master
pachctl list file returns@master
```
You should see the following files in each repo:
- Purchases:

    ![stores_repository](./img/pachctl_list_file_stores_master.png)

- Stores:

    ![purchases_repository](./img/pachctl_list_file_purchases_master.png)

- Returns:

    ![returns_repository](./img/pachctl_list_file_returns_master.png)
 
 You are ready to run any of the following examples.
## 4. Example 1 : An Inner Join pipeline 

***Goal***

We want to list all purchases made in stores sharing the same zip code. 

Following the 2 steps pattern described in our [Datum processing](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/) documentation, we will need 2 pipelines:

- One that performs an **inner join** on the 'purchases' and 'stores' repos then output one text file per matched datum. We designed the source code of the pipeline so that it sorts those text files into directories named after the zip code of the store in which a purchase was made.

    1. **Pipeline input repositories**: The `purchases` and `stores` repos are joined (`inner join`) on the STOREID of their file name.

        | Purchases | Stores |
        |-----------|--------|
        | ![stores_repository](./img/pachctl_list_file_stores_master.png)|![purchases_repository](./img/pachctl_list_file_purchases_master.png)|

    2. **Pipeline**: The [inner_join.json](./pipelines/inner/inner_join.json) pipeline will first create input datums following its `glob` pattern and `"join_on": "$1"` values (see the pipeline specification), then execute some Python code reading the `zipcode` in the STOREIDx.txt file of each datum.

    3. **Pipeline output repository**: The output repo  `inner_join` will contain a list of directories named after the zipcodes. Each will contain a text file per matched datum.
    We will expect the 2 following directories in the output repository of our `inner_join` pipeline: 
        ```
            └── /02108/
            └── /90210/
        ```

- A following pipeline will then "glob" all files in each "zipcode" directory then merge their content into a final text file.
    1. **Pipeline input repositories**: The `inner_join` output repo becomes the input repo of the following pipeline.

    2. **Pipeline**: The [reduce_inner.json](./pipelines/inner/reduce_inner.json) pipeline "globs" each directory (`"glob": "/*"`) and merge the content of all their file into one txt file name after the directory name (zipcode).

    3. **Pipeline output repository**: The output repo `reduce_inner` will contain 2 text files. Each will list all purchases made in a specific zip code. 

        We will expect the 2 following files in the output repository of the pipeline: 
        ```shell
            └── 02108.txt
            └── 90210.txt
        ```
        Each should contain 3 purchases.

While going through the process, we will understand what datums are created by our `inner join`. 


***Step 1*** - Before we create our `inner join` pipeline, let's preview what our datums will look like by running the following command in the `examples/joins` directory:

```shell
pachctl list datum -f pipelines/inner/inner_join.json
```
![list_datum_outer_inner](./img/pachctl_list_datum_inner.png)

Note that:

- 6 datums are created corresponding to the 6 purchases in the `purchases` repo. 
- There was no purchase at the STOREID4; therefore **no datum was created** for this Store ID. See diagram below. This is a characteristic of an `inner join`. **You only see the stores where purchases were made** (a datum is created if, and only if, there is a match).

![inner_join_list_datum](./img/inner_join_list_datum.png)

***Step 2*** - Let's now create our pipelines and see them in action. 

Because unprocessed data are awaiting in our input repositories, the creation of the pipelines will automatically trigger 2 jobs.
In the `examples/joins` directory, run:
```shell
pachctl create pipeline -f pipelines/inner/inner_join.json
pachctl create pipeline -f pipelines/inner/reduce_inner.json
```

Check that your pipelines ran successfully:
```shell
pachctl list pipeline
```

***Step 3*** - Check the output repository of the first pipeline. Our code has created one txt file per datum. Each datum has been placed in a directory named after the zip code of its store: 

```shell
pachctl list file inner_join@master
```
You should see 2 directories, each containing 3 text files.

![output_repository](./img/pachctl_list_file_inner_join_master.png)

For visual confirmation of their content, run:

```shell
pachctl list file inner_join@master:/02108/
```
The following table lists the expected results for each zipcode:

|/02108/|/90210/|
|----------|----------|
/02108/ORDERW078929_STOREID2.txt file 104B|/90210/ORDERW080231_STOREID5.txt file 102B|
/02108/ORDERW080520_STOREID1.txt file 105B|/90210/ORDERW080528_STOREID5.txt file 102B|
/02108/ORDERW080521_STOREID1.txt file 105B|/90210/ORDERW598471_STOREID3.txt file 104B|

Now check the output repo of the second pipeline:

```shell
pachctl list file reduce_inner@master
```
and notice the 2 text files `/02108.txt` and
`/90210.txt`. 

you can also check the content of each file by running:

```shell
pachctl get file  reduce_inner@master:/02108.txt
```
Find the detail of what you should see in the table below:

|/02108.txt|/90210.txt|
|----------|----------|
|Purchase at store: 1  ... ORDER W080520|Purchase at store: 3 ... ORDER W598471|
|Purchase at store: 1 ... ORDER W080521|Purchase at store: 5 ... ORDER W080528| 
|Purchase at store: 2 ... ORDER W078929 |Purchase at store: 5 ... ORDER W080231| 

>![pach_logo](./img/pach_logo.svg) Want to take this example to the next level? Practice using [groups](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/group/) AND joins. You can create a multi-step pipeline that will group Returns and Purchases by storeID then join the output repo with Stores to aggregate by location. 

## 5. Example 2 : Outer Join pipeline creation 
>![pach_logo](./img/pach_logo.svg) You specify an [outer join](#case-1-outer-join-on-the-returns-repo-only) by adding an "outer_join" boolean property to an input repo in the `join` section of your pipeline spec. 

***Goal***

We are going to list all returns by zip code and understand the datums created by our `outer join` when setting the `"outer_join": true` on:

1. The `returns` repository only
2. The `stores` repository only
3. Both our repos `stores` and `returns` repositories

These 3 cases will create 3 different sets of datums that we will explain further. 

Following the 2 steps pattern described in our [Datum processing](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/) documentation, we will need 2 pipelines:

- One that performs an **outer join** on the 'returns' and 'stores' repos then outputs one text file per matched datum. We designed the source code of the pipeline so that it sorts those text files into directories named after the zip code of the store in which a purchase was made. 

    1. **Pipeline input repositories**: The `stores` and `returns` repos are joined (See 3 cases of outer join above) on the STOREID of their filename.
    | Returns | Stores | |-----------|--------| |![returns_repository](./img/pachctl_list_file_returns_master.png)|![purchases_repository](./img/pachctl_list_file_purchases_master.png)|
    
    2. **Pipeline**: Our [outer_join.json](./pipelines/outer/outer_join.json) pipeline will first create input datums following its `glob` pattern, `"join_on": "$1"`, and `"outer_join"` values, then execute some Python code reading the `zipcode` in the STOREIDx.txt file of each datum (if any).
    
    3. **Pipeline output repository**: The output repo `outer_join` will contain a list of directories named after the zipcodes. Each will include a text file per matched datum. Expect different results depending on where the outer_join was set.

- A following pipeline will then "glob" all files in each "zipcode" directory then merge their content into a final text file.

    1. **Pipeline input repositories**: The `outer_join` output repo becomes the input repo of the following pipeline.

    2. **Pipeline**: The [reduce_outer.json](./pipelines/outer/reduce_outer.json) pipeline "globs" each directory (`"glob": "/*"`) and merge the content of all their file into one txt file name after the directory name (zipcode).

    3. **Pipeline output repository**: The output repo `reduce_outer` will contain a list of text files named after the stores' zip code. The detail of each will depend on where the 'outer join' was put.

While going through the process, we will understand what datums are created by our 3 options of `outer join`. 

We have listed all the possible outcomes in the following cheat sheet. Each particular case will be further explained:
![outer_join_list_datum](./img/outer_joins_list_datum.png)


### ***Case 1*** Outer join on the Returns repo only
1. In the `examples/joins/pipelines/outer` directory, edit the pipeline's specification `outer_join.json` and set `"outer_join" :true` on the repo `returns`:
    ```shell
    "input": {
        "join": [
        {
            "pfs": {
            "repo": "stores",
            "branch": "master",
            "glob": "/STOREID(*).txt",
            "join_on": "$1",
            "outer_join": false
            }
        },
        {
        "pfs": {
            "repo": "returns",
            "branch": "master",
            "glob": "/*_STOREID(*).txt",
            "join_on": "$1",
            "outer_join": true
        }
        }
    ]
    },
    ```
1. Before we create our `outer join` pipeline, let's preview the datums that will be created. 
    In the same directory, run:

    ```shell
    pachctl list datum -f outer_join.json
    ```
    ![list_datum_outer_returns](./img/pachctl_list_datum_outer_returns.png)
 
 
    What you are noticing is that all your 4 returns are showing in a datum. 
    >![pach_logo](./img/pach_logo.svg) Note, however, that one return was made in a store **not** listed in our repo (STOREID0). There was no match on any STOREID for this specific return file, however; **a datum was created**, containing only the return file. By setting the returns repo as `"outer_join":true`, you are requesting to **see ALL of the repo's files reflected in datums, whether there is a match or not.

1. In the same directory, let's now create our pipelines by running:

    ```shell
    pachctl create pipeline -f outer_join.json
    pachctl create pipeline -f reduce_outer.json
    ```
1. Check that your pipelines ran successfully:

    ```shell
    pachctl list pipeline
    ```

1. Check the output repository of the first pipeline. 
    Our code has created one txt file per datum. Each datum has been placed in a directory named after the zipcode of its store: 

    ```shell
    pachctl list file outer_join@master
    ```
    You should see 3 directories.

    ```
    /02108/   dir  103B
    /90210/   dir  200B
    /UNKNOWN/ dir  99B
    ```

1. Now check the output repo of the second pipeline:

    ```shell
    pachctl list file reduce_outer@master
    ```
    Notice that an UNKNOWN.txt file shows up listing the return without matching store.

    ![](./img/pachctl_list_file_outer_returns.png)

    You can also check the content of each file by running:

    ```shell
    pachctl get file  reduce_outer@master:/02108.txt
    ```

    The following table lists the expected results for this scenario:

    |/02108.txt|/90210.txt|/UNKNOWN.txt|
    |----------|----------|------------|
    |Return at store: 1 ... ORDER W080520|Return at store: 5 ... ORDER W080528|This return store does not exist ... ORDER W261452 |
    | | Return at store: 5 ... ORDER W080231| |
    | | | |


### ***Case 2*** Outer join on the Stores repo only
1. In the `examples/joins/pipelines/outer` directory, set `"outer_join" :true` on the `stores` repo in `outer_join.json`:

    ```shell
    "input": {
        "join": [
        {
            "pfs": {
            "repo": "stores",
            "branch": "master",
            "glob": "/STOREID(*).txt",
            "join_on": "$1",
            "outer_join": true
            }
        },
        {
        "pfs": {
            "repo": "returns",
            "branch": "master",
            "glob": "/*_STOREID(*).txt",
            "join_on": "$1",
            "outer_join": false
        }
        }
    ]
    },
    ```
1. Preview all the datums again: 

    In the same directory, run:
    ```shell
    pachctl list datum -f outer_join.json
    ```
    ![list_datum_outer_returns](./img/pachctl_list_datum_outer_stores.png)
 
 
    >![pach_logo](./img/pach_logo.svg) Note that **all** of the stores files are showing up in the datums: some are associated with a matched return, some are not, yet **a datum was created**. Additionally, the return made to STOREID0 did **not** generate a datum since there is no STOREID0.txt in the Stores repo.

1. Update your pipeline. In the `examples/joins/pipelines/outer` directory, run:

    ```shell
    pachctl update pipeline -f outer_join.json --reprocess
    ```

1. Check the output repository of your last pipeline and notice that a 94107.txt file now shows up (STOREID4 has a different zip code 94107).

    ```shell
    pachctl list file reduce_outer@master
    ```
    ![](./img/pachctl_list_file_outer_stores.png)

1. Visualize their content:

    ```shell
    pachctl get file reduce_outer@master:/02108.txt
    ```
    Expected results for this scenario:

    |/02108.txt|/90210.txt|/94107.txt|
    |----------|----------|------------|
    |Return at store: 2 ... NONE|Return at store: 5 ... ORDER W080231|Return at store: 4 ... NONE|
    |Return at store: 1 ... ORDER W080520| Return at store: 3 ... NONE| |
    | |Return at store: 5 ... ORDER W080528| |

### ***Case 3*** Outer join on both the Stores and Returns repos
1. Set `"outer_join" :true` on both repos `returns` and `stores` in `outer_join.json`.
1. Preview your datums then create/update your pipeline: 

    ![list_datum_full_outer](./img/pachctl_list_datum_full_outer.png)
 
 
    >![pach_logo](./img/pach_logo.svg) Note that in this last case, the datums contain all of the returns (with or without a matching store) **and** all of the stores (with or without a matching return).


1. Check the output repository of your pipeline and notice that all the zipcode.txt files now show up, including the UNKNOWN.txt.

    ![](./img/pachctl_list_file_full_outer.png)

1. Expected results for this scenario:

    |/02108.txt|/90210.txt|/94107.txt|UNKNOWN.txt|
    |----------|----------|------------|-----------|
    |Return at store: 2 ... NONE|Return at store: 5 ... ORDER W080231|Return at store: 4 ... NONE| This return store does not exist ... ORDER W261452|
    |Return at store: 1 ... ORDER W080520| Return at store: 3 ... NONE| | |
    | |Return at store: 5 ... ORDER W080528| | |


>![pach_logo](./img/pach_logo.svg) **Side note**: We could also aggregate the files by STOREID using a [group input](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/group/). In that case, the datums produced would be bigger. Each datum would contain one STOREIDx.txt file and ALL of the matching files for this given storeID in the grouped repos. The choice between group and join is an architectural choice to optimize your pipelines' parallelization and processing time.
       
We encourage you to run our [group example](https://github.com/pachyderm/pachyderm/tree/master/examples/group) to understand more of the differences between join and group. Look for the `retail_group.json` pipeline.

You can also follow those quick steps:
```shell
cd ../../../group
pachctl list datum -f retail_group.json
```

In this last pipeline, we have grouped `purchases`, `returns`, and `stores` repos on STOREID. Look at the datums created.

![pach_logo](./img/pachctl_list_datum_retail_group.png)


