# Quick Start Guide: Web Scraper
In this guide you're going to create a Pachyderm pipeline to scrape web pages. 
We'll use a standard unix tool, `wget`, to do our scraping.

## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster. [Installation instructions can be found here](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html).

## Create a Repo

A `Repo` is the highest level primitive in `pfs`. Like all primitives in pfs, they share
their name with a primitive in Git and are designed to behave analogously.
Generally, a `repo` should be dedicated to a single source of data such as log
messages from a particular service. Repos are dirt cheap so don't be shy about
making them very specific.

For this demo we'll simply create a `repo` called
“urls” to hold a list of urls that we want to scrape.

```shell
$ pachctl create-repo urls
$ pachctl list-repo
urls
```


## Start a Commit
Now that we’ve created a `repo` we’ve got a place to add data.
If you try writing to the `repo` right away though, it will fail because you can't write directly to a
`Repo`. In Pachyderm, you write data to an explicit `commit`. Commits are
immutable snapshots of your data which give Pachyderm its version control for
data properties. Unlike Git though, commits in Pachyderm must be explicitly
started and finished.

Let's start a new commit in the “urls” repo:
```shell
$ pachctl start-commit urls master
master/0
```

This returns a brand new commit id. Yours should be different from mine.
Now if we take a look inside our repo, we’ve created a directory for the new commit:
```shell
$ pachctl list-commit urls
master/0
```

A new directory has been created for our commit and now we can start adding
data. Data for this example is just a single file with a list of urls. We've provided a sample file for you with just 3 urls, Google, Reddit, and Imgur.
We're going to write that data as a file called “urls” in pfs.

```shell
# Write sample data into pfs
$ cat examples/scraper/urls | pachctl put-file urls master/0 urls
```

## Finish a Commit

Pachyderm won't let you read data from a commit until the `commit` is `finished`.
This prevents reads from racing with writes. Furthermore, every write
to pfs is atomic. Now let's finish the commit:

```shell
$ pachctl finish-commit urls master/0
```

Now we can view the file:

```shell
$ pachctl get-file urls master/0 urls
www.google.com
www.reddit.com
www.imgur.com
```
However, we've lost the ability to write to this `commit` since finished
commits are immutable. In Pachyderm, a `commit` is always either _write-only_
when it's been started and files are being added, or _read-only_ after it's
finished.

## Create a Pipeline

Now that we've got some data in our `repo` it's time to do something with it.
Pipelines are the core primitive for Pachyderm's processing system (pps) and
they're specified with a JSON encoding. We're going to create a pipeline that simply scrapes each of the web pages in “urls.”

```
+----------+     +---------------+
|input data| --> |scrape pipeline|
+----------+     +---------------+
```

The `pipeline` we're creating can be found at [scraper.json](scraper.json).  The full content is also below.
```
{
  "pipeline": {
    "name": "scraper”
  },
  "transform": {
    "cmd": [ "wget",
        "--recursive",
        "--level", "1",
        "--accept", "jpg,jpeg,png,gif,bmp",
        "--page-requisites",
        "--adjust-extension",
        "--span-hosts",
        "--no-check-certificate",
        "--timestamping",
        "--directory-prefix",
        "/pfs/out",
        "--input-file", "/pfs/urls/urls"
    ],
    "acceptReturnCode": [4,5,6,7,8]
  },
  "parallelism": "1",
  "inputs": [
    {
      "repo": {
        "name": "urls"
      }
    }
  ]
}
```

In this pipeline, we’re just using `wget` to scrape the content of our input web pages. “level” indicates how many recursive links `wget` will retrieve. We currently have it set to 1, which will only scrape the home page, but you can crank it up later if you want.

Another important section to notice is that we read data
from `/pfs/urls/urls` (/pfs/[input_repo_name]) and write data to `/pfs/out/`.  We create a directory for each url in “urls” with all of the relevant scrapes as files.

Now let's create the pipeline in Pachyderm:

```shell
$ pachctl create-pipeline -f doc/examples/scraper/scraper.json
```

## What Happens When You Create a Pipeline
Creating a `pipeline` tells Pachyderm to run your code on *every* finished
`commit` in a `repo` as well as *all future commits* that happen after the pipeline is
created. Our `repo` already had a `commit` with the file “urls” in it so Pachyderm will automatically
launch a `job` to scrape those webpages.

You can view the job with:

```shell
$ pachctl list-job
ID                                 OUTPUT                                     STATE
09a7eb68995c43979cba2b0d29432073   scraper/2b43def9b52b4fdfadd95a70215e90c9   JOB_STATE_RUNNING
```

Depending on how quickly you do the above, you may see `running` or
`success`.

Pachyderm `job`s are implemented as Kubernetes jobs, so you can also see your job with:

```shell
$ kubectl get job
JOB                                CONTAINER(S)   IMAGE(S)             SELECTOR                                                         SUCCESSFUL
09a7eb68995c43979cba2b0d29432073   user           pachyderm/job-shim   app in (09a7eb68995c43979cba2b0d29432073),suite in (pachyderm)   1
```

Every `pipeline` creates a corresponding `repo` with the same
name where it stores its output results. In our example, the pipeline was named “scraper” so it created a `repo` called “scraper” which contains the final output.


## Reading the Output
There are a couple of different ways to retrieve the output. We can read a single output file from the “scraper” `repo` in the same fashion that we read the input data:

```shell
$ pachctl list-file scraper 2b43def9b52b4fdfadd95a70215e90c9 urls
$ pachctl get-file scraper 2b43def9b52b4fdfadd95a70215e90c9 urls/www.imgur.com/index.html
```

Using `get-file` is good if you know exactly what file you’re looking for, but for this example we want to just see all the scraped pages. One great way to do this is to mount the distributed file system locally and then just poke around.

## Mount the Filesystem
First create the mount point:

```shell
$ mkdir ~/pfs
```

And then mount it:

```shell
# We background this process because it blocks.
$ pachctl mount ~/pfs &
```

This will mount pfs on `~/pfs` you can inspect the filesystem like you would any
other local filesystem. Try:

```shell
$ ls ~/pfs
urls
scraper
```
You should see the urls repo that we created.

Now you can simply `ls` and `cd` around the file system. Try pointing your browser at the scraped output files!


## Processing More Data

Pipelines can be triggered manually, but also will automatically process the data from new commits as they are
created. Think of pipelines as being subscribed to any new commits that are
finished on their input repo(s).

If we want to re-scrape some of our urls to see if the sites of have changed, we can use the `run-pipeline` command:

```shell
$ pachctl run-pipeline scraper
fab8c59c786842ccaf20589e15606604
```

Next, let’s add additional urls to our input data . We're going to append more urls from “urls2” to the file “urls.”

We first need to start a new commit to add more data. Similar to Git, commits have a parental
structure that track how files change over time. Specifying a parent is
optional when creating a commit (notice we didn't specify a parent when we
created the first commit), but in this case we're going to be adding
more data to the same file “urls.”


Let's create a new commit with our previous commit as the parent:

```shell
$ pachctl start-commit urls master
master/1
```

Append more data to our urls file in the new commit:
```shell
$ cat examples/scraper/urls2 | pachctl put-file urls master/1 urls
```
Finally, we'll want to finish our second commit. After it's finished, we can
read “scraper” from the latest commit to see all the scrapes.

```shell
$ pachctl finish-commit urls master1
```
Finishing this commit will also automatically trigger the pipeline to run on
the new data we've added. We'll see a corresponding commit to the output
“scraper” repo with data from our newly added sites.

```shell
$ pachctl list-commit scraper
```
## Next Steps
You've now got a working Pachyderm cluster with data and a pipelines! Here are a few ideas for next steps that you can expand on your working setup.
- Add a bunch more urls and crank up the “level” in the pipeline. You’ll have to delete the old pipeline and re-create or give your pipeline and new name.
- Add a new pipeline than does something interesting with the scraper output. Image or text processing could be fun. Just create a pipeline with the scraper repo as an input.

We'd love to help and see what you come up with so submit any issues/questions you come across or email at info@pachyderm.io if you want to show off anything nifty you've created!
