**Note**: This is a Pachyderm pre version 1.4 tutorial.  It needs to be updated for the latest versions of Pachyderm.

# Beginner Tutorial: Fruit Stand

In this guide you're going to create a Pachyderm pipeline to process transaction logs from a fruit stand. We'll use standard unix tools `bash`, `grep` and `awk` to do our processing. Thanks to Pachyderm's processing system we'll be able to run the pipeline in a distributed, streaming fashion. As new data is added, the pipeline will automatically process it and materialize the results.

If you hit any errors not covered in this guide, check our [troubleshooting](http://docs.pachyderm.io/en/v1.7.3/managing_pachyderm/general_troubleshooting.html) docs for common errors, submit an issue on [GitHub](https://github.com/pachyderm/pachyderm), join our [users channel on Slack](http://pachyderm-users.slack.com), or email us at [support@pachyderm.io](mailto:support@pachyderm.io) and we can help you right away.

## Prerequisites

This guide assumes that you already have Pachyderm running locally. Check out our [local installation](http://docs.pachyderm.io/en/v1.7.3/getting_started/local_installation.html) instructions if haven't done that yet and then come back here to continue.


## Create a Repo

A `repo` is the highest level primitive in the Pachyderm file system (pfs). Like all primitives in pfs, it shares its name with a primitive in Git and is designed to behave somewhat analogously. Generally, repos should be dedicated to a single source of data such as log messages from a particular service, a users table, or training data for an ML model. Repos are dirt cheap so don't be shy about making tons of them.

For this demo, we'll simply create a repo called
"data" to hold the data we want to process:

```sh
$ pachctl create-repo data

# See the repo we just created
$ pc list-repo
NAME CREATED       SIZE 
data 8 seconds ago 0B   
```

## Adding Data to Pachyderm

Now that we've created a repo it's time to add some data. In Pachyderm, you write data to a `commit` (again, similar to Git). Commits are immutable snapshots of your data which give Pachyderm its version control properties. `Files` can be added, removed, or updated in a given commit and then you can view a diff of those changes compared to a previous commit.

Let's start by just adding a file to a new commit. We've provided a sample data file for you to use in our GitHub repo -- it's a list of purchases from a fruit stand.

We'll use the `put-file` command along with a flag, `-f`. `-f` can take either a local file or a URL, and in our case, we'll pass the sample data on GitHub.

 We also specificy the repo name `data`, the branch name `master`, and the path `/sales` to which the data will be written  (under the hood, `put-file` creates a new commit in whatever branch it's passed).

```sh
$ pachctl put-file data master /sales -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.7.3/examples/fruit_stand/set1.txt
```

Finally, we can see the data we just added to Pachyderm.
```sh
# If we list the repos, we can see that there is now data
$ pc list-repo
NAME CREATED        SIZE 
data 31 seconds ago 874B 

# We can view the commit we just created
$ pc list-commit data
REPO ID                               PARENT STARTED        DURATION           SIZE 
data b60d583b2d2747b6ae912c1e7fe1fb06 <none> 24 seconds ago Less than a second 874B 

# We can also view the contents of the file that we just added
$ pachctl get-file data master /sales | head -n5
orange	4
banana	2
banana	9
orange	9
apple	6
```

## Create a Pipeline

Now that we've got some data in our repo, it's time to do something with it. `Pipelines` are the core primitive for Pachyderm's processing system (pps) and they're specified with a JSON encoding. For this example, we've already created two pipelines for you, which can be found at [examples/fruit_stand/pipeline.json on Github](https://github.com/pachyderm/pachyderm/blob/v1.7.3/examples/fruit_stand/pipeline.json). Please open a new tab to view the pipeline while we talk through it.

When you want to create your own pipelines later, you can refer to the full [pipeline spec](http://docs.pachyderm.io/en/v1.7.3/reference/pipeline_spec.html) to use more advanced options. This includes building your own code into a container instead of just using simple shell commands as we're doing here.

For now, we're going to create two pipelines to process this hypothetical sales data. The first filters the sales logs into separate records for apples, oranges and bananas. The second step sums these sales numbers into a final sales count.

```
 +----------+     +--------------+     +------------+
 |input data| --> |filter pipline| --> |sum pipeline|
 +----------+     +--------------+     +------------+
```

In the first step of this pipeline, we are grepping for the terms "apple", "orange", and "banana" and writing that line to the corresponding file. Notice we read data from `/pfs/data` (in general, `/pfs/[input_repo_name]`) and write data to `/pfs/out/`. These are special local directories that Pachyderm creates within the container for you. All the input data will be found in `/pfs/[input_repo_name]` and output data should always go in `/pfs/out`.

The second step of this pipeline takes each file, removes the fruit name, and sums up the purchases. The output of our complete pipeline is three files, one for each type of fruit with a single number showing the total quantity sold.

Now let's create the pipeline in Pachyderm:

```sh
$ pachctl create-pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.7.3/examples/fruit_stand/pipeline.json
```

## What Happens When You Create a Pipeline

Creating a pipeline tells Pachyderm to run your code on the most recent input commit as all future commits that happen after the pipeline is created. Our repo already had a commit, so Pachyderm automatically launched a `job` to process that data.

You can view the job with:

```sh
 $ pc list-job
ID                               OUTPUT COMMIT                           STARTED        DURATION  RESTART PROGRESS  DL   UL   STATE            
f8596ec0628a4ed6846489971b5c75e3 sum/0069dd5ed4644a7d87c7092204e7d3cf    21 seconds ago 5 seconds 0       3 + 0 / 3 200B 12B  success 
b26a1c64aece4d00a66ec3892735e474 filter/e1482b25ed574005919f325b67e3fc8d 21 seconds ago 5 seconds 0       1 + 0 / 1 874B 200B success 
```

Every pipeline creates a corresponding output repo with the same name as the pipeline itself, where it stores its results. In our example, the "filter" transformation created a repo called "filter" which was the input to the "sum" transformation. The "sum" repo contains the final output files.

```sh
$ pc list-repo
NAME   CREATED            SIZE 
sum    About a minute ago 12B  
filter About a minute ago 200B 
data   3 minutes ago      874B 
```

## Reading the Output

 We can read the output data from the "sum" repo in the same fashion that we read the input data:

```
$ pachctl get-file sum master /apple
133
```

## Processing More Data

Pipelines will also automatically process the data from new commits as they are created (think of pipelines as being subscribed to any new commits that are finished on their input branches). Also similar to Git, commits have a parental structure that track how files change over time.

In this case, we're going to be adding more data to the file "sales". Our fruit stand business might append to this file every every hour with all the new purchases that happened in that window.

Let's create a new commit with our previous commit as the parent and add more sample data (set2.txt) to "sales":

```sh
$ pachctl put-file data master sales -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.7.3/examples/fruit_stand/set2.txt
```

Adding a new commit of data will automatically trigger the pipeline to run on the new data we've added. We'll see a corresponding commit to the output "sum" repo with files "apple", "orange" and "banana" each containing the cumulative total of purchases. Let's read the "apples" file again and see the new total number of apples sold.

```sh
$ pc list-commit data
REPO ID                               PARENT                           STARTED        DURATION           SIZE     
data 2fbfec34ab534f869c343654a8c1977b b60d583b2d2747b6ae912c1e7fe1fb06 1 minute ago   Less than a second 1.696KiB 
data b60d583b2d2747b6ae912c1e7fe1fb06 <none>                           1 minute ago   Less than a second 874B     

$ pc list-job
ID                               OUTPUT COMMIT                           STARTED        DURATION           RESTART PROGRESS  DL       UL   STATE            
8d1c660acaed42c58c37e9c538f943cc sum/79490b7548ed4dd881e3d4096e4fe553    54 seconds ago Less than a second 0       3 + 0 / 3 400B     12B  success 
4f281923c38549aeaf88f68707510368 filter/ce9a565a63e94b618fc6f8c78d148779 54 seconds ago Less than a second 0       1 + 0 / 1 1.696KiB 400B success 
f8596ec0628a4ed6846489971b5c75e3 sum/0069dd5ed4644a7d87c7092204e7d3cf    8 minutes ago  5 seconds          0       3 + 0 / 3 200B     12B  success 
b26a1c64aece4d00a66ec3892735e474 filter/e1482b25ed574005919f325b67e3fc8d 8 minutes ago  5 seconds          0       1 + 0 / 1 874B     200B success 

$ pachctl get-file sum master /apple
324
```

## Next Steps
You've now got Pachyderm running locally with data and a pipeline! If you want to keep playing with Pachyderm locally, here are some ideas to expand on your working setup.

  - Write a script to stream more data into Pachyderm. We already have one in Golang for you on [GitHub](https://github.com/pachyderm/pachyderm/tree/v1.7.3/examples/fruit_stand/generate) if you want to use it.
  - Add a new pipeline that does something interesting with the "sum" repo as an input.
  - Add your own data set and `grep` for different terms. This example can be generalized to generic word count.

You can also start learning some of the more advanced topics to develop analysis in Pachyderm:

- [Deploying on the cloud](http://docs.pachyderm.io/en/v1.7.3/deployment/deploy_intro.html)
- [Input data from other sources](http://docs.pachyderm.io/en/v1.7.3/fundamentals/getting_data_into_pachyderm.html)
- [Create pipelines using your own code](http://docs.pachyderm.io/en/v1.7.3/fundamentals/creating_analysis_pipelines.html)

We'd love to help and see what you come up with so submit any issues/questions you come across on [GitHub](https://github.com/pachyderm/pachyderm) , [Slack](http://pachyderm-users.slack.com) or email at dev at pachyderm.io if you want to show off anything nifty you've created!
