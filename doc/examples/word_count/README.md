# Pachyderm Word Count

In this guide, we will write a classic [word count](https://portal.futuresystems.org/manual/hadoop-wordcount) application on Pachyderm.  This is a somewhat advanced guide; to learn the basic usage of Pachyderm, start with the [beginner tutorial](http://pachyderm.readthedocs.io/en/stable/getting_started/beginner_tutorial.html).

## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster. [Installation instructions can be found here](http://pachyderm.readthedocs.io/en/stable/getting_started/local_installation.html).

## Pipelines

In this example, we will have three processing stages defined by three pipeline stages.  We say that these pipeline stages are "connected" because one pipeline's output is the input of the other.

## input

Let's create the first pipeline:

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f input_pipeline.json
```

This first pipeline, `input`, uses `wget` to download web pages from Wikipedia which will be used as the input for the next pipeline.  We set `parallelism` to 1 because `wget` can't be parallelized.

Note how this pipeline itself has no inputs.  A pipeline that doesn't have inputs can only be triggered manually:

```
$ pachctl run-pipeline input
```

Now you should be able to see a job running:

```
$ pachctl list-job
```

## map

This pipeline counts the number of occurrences of each word it encounters.  While this task can very well be accomplished in bash, we will demonstrate how to use custom code in Pachyderm by using a [Go program](map.go).

First of all, we build a [Docker image](Dockerfile) that contains our Go program:

```
$ docker build -t wordcount-map .
```

Then, we simply refer to the image in our pipeline and run the pipeline:

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f mapPipeline.json
```

As soon as you create this pipeline, it will start processing data from `input`.  To parallelize the pipeline, you can modify the parallelism spec, `mapPipeline.json`. Pachyderm will then spin up concurrently running containers where each container gets `1/N` of the data (where N in the number of containers).  For each word a container encounters, it writes a file whose filename is the word, and whose content is the number of occurrences.  If multiple containers write the same file, the content is merged.  As an example, the file `morning` might look like this:

```
# find the commit ID
$ pachctl list-job -p map  
ID                                 OUTPUT                                           STARTED             DURATION            STATE
6600c71be4e8604f716ce1965895dc27   map/5b0e5c8f345b4f8c9690c77333564687   46 hours ago        -                   success
$ pachctl get-file map 5b0e5c8f345b4f8c9690c77333564687 morning 
217
355
142
```

This shows that there were three containers that wrote to the file `morning`.

## reduce

The final pipeline goes through every file and adds up the numbers in each file.  For this pipeline we can use a simple bash script:

```
find /pfs/map -name '*' | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count /pfs/out/`basename $count`; done
```

Which we bake into [reducePipeline.json](./reducePipeline.json).

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f reducePipeline.json
```

The final output might look like this:

```
$ pachctl get-file reduce [commit ID] morning
714
```

To get a complete list of the words counted:

```
$ pachctl list-file reduce [commit ID]
```


## Preliminary Benchmarks

We ran this pipeline on 7.5GB of data in July 2016, on a 3-node GCE cluster with 4 CPUs and 15GB of memory each.  The job completed in 9 hours.
