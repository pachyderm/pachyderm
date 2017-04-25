# Pachyderm Word Count

In this guide, we will write a classic [word count](https://portal.futuresystems.org/manual/hadoop-wordcount) application on Pachyderm.  This is a somewhat advanced guide; to learn the basic usage of Pachyderm, start with the [beginner tutorial](http://pachyderm.readthedocs.io/en/stable/getting_started/beginner_tutorial.html).

## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster. [Installation instructions can be found here](http://pachyderm.readthedocs.io/en/stable/getting_started/local_installation.html).

## Pipelines

In this example, we will have three processing stages defined by three pipeline stages.  We say that these pipeline stages are "connected" because one pipeline's output is the input of the other.

Our first pipeline is a web scraper that just pulls content from the internet. Our second pipeline does a "map" step to tokenize the words from the scraped pages. Our final step is a "reduce" to aggreate the the totals. 

All three pipelines, including the "reduce", can be run in a distributed fashion to maximize performance. 

## Input

Our input data is a set of files. Each file is named for the site we want to scrape with the content being the URL. 

Let's create the input repo and add one URL, Wikipedia:
```
$ pachctl create-repo urls

# We assume you're running this from the root of this repo:
$ pachctl put-file urls master -c -f doc/examples/word_count/Wikipedia
```

Let's create the first pipeline:

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/wordcount_scraper.json
```

This first pipeline, `scraper`, uses `wget` to download web pages from Wikipedia which will be used as the input for the next pipeline.


Now you should be able to see a job running and a new repo called `scraper` that contains the output of our scrape. 

```
$ pachctl list-job

$ pachctl list-repo
```

The output of our scraper pipeline has a file structure like:

```
Wikipedia
 |--/page1
 |--/page2
 |--/page3
```

## Map

This pipeline counts the number of occurrences of each word it encounters.  While this task can very well be accomplished in bash, we will demonstrate how to use custom code in Pachyderm by using a [Go program](map.go).

We need to build a Docker image that contains our Go program. We can do it ourselves using the provided [Dockerfile](Dockerfile). 

```
$ docker build -t wordcount-map.
```

This builds the image locally. You'll need to push the image to a registry that Pachyderm can access. Either DockerHub, your own internal registry, or you can build it inside Minikube if you're working locally. 

The `image` field in our pipeline spec, mapPipeline.json, simply needs to point to the right location for the image. `mapPipeline.json` (shown below) references a locally built image as if you built it within Minikube. 

If you don't want to build this image yourself and add it to a registry, you can just reference our public image on dockerhub by changing the image field to:

```
 "image": "pachyderm/wordcount-map"
```

Now let's create the Map pipeline. 
```
# Again, we assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/mapPipeline.json
```

As soon as you create this pipeline, it will start processing data from `scraper`. For each web page the map.go code processes, it writes a file a file for each encountered word whose filename is the word, and whose content is the number of occurrences.  If multiple workers write to the same file, the content is concatenated.  As an example, the file `morning` might look like this:

```
$ pachctl get-file map master morning 
36
11
17
```

This shows that there were three [datums](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html)(websites) that included to the word and wrote to the file `morning`.

## Reduce

The final pipeline goes through every file and adds up the numbers in each file.  For this pipeline we can use a simple bash script:

```
find /pfs/map -name '*' | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count /pfs/out/`basename $count`; done
```

Which we bake into [reducePipeline.json](reducePipeline.json).

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/reducePipeline.json
```

The final output might look like this:

```
$ pachctl get-file reduce master morning
64
```

To get a complete list of the words counted:

```
$ pachctl list-file reduce master
```

## Expand on the example

Now that we've got a full end-to-end scraper and wordcount use case set up, lets add more to it. First, let's add more data. Go ahead and add a few more sites to scrape. 

```
# Instead of using the -c shorthand flag, let's do this the long way by starting a commit, adding files, and then finishing the commit.
$ pachctl start-commit urls master

# Reminder: files added should be named for the website and have the URL as the content. You'll have to create these files.
$ pachctl put-file urls master -f HackerNews
$ pachctl put-file urls master -f Reddit
$ pachctl put-file urls master -f GitHub

$ pachctl finish-commit urls master
```
Your scraper should automatically get started pulling these new sites (it won't rescrape Wikipedia). That'll automatically trigger the Map and Reduce pipelines to process the new data too and update the word counts for all the sites combined.

If you add a bunch more data and your pipeline starts to run slowly, you can crank up the parallism. By default, pipelines spin up one worker for each node in your cluster, but you can set that manually with the [parallelism spec](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#parallelism-spec-optional). The pipeline are already configured to spread computation across the various workers with `"glob": "/*". Check out our [distributed computing docs](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html) to learn more about that. 


