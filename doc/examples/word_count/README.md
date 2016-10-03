# Quick Start Guide: Word Count

In this guide, we will write a classic [word count](https://portal.futuresystems.org/manual/hadoop-wordcount) application on Pachyderm.  This is a somewhat advanced guide; to learn the basic usage of Pachyderm, start with the [beginner tutorial](http://pachyderm.readthedocs.io/en/latest/getting_started/beginner_tutorial.html).

## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster. [Installation instructions can be found here](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html).

## Pipelines

In this example, we will have three connected pipelines.  We say that two pipelines are "connected" if one pipeline's output is the input of the other.

## wordcount_input

Let's create the first pipeline:

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/input_pipeline.json
```

This first pipeline, `wordcount_input`, uses `wget` to download web pages from Wikipedia which will be used as the input for the next pipeline.  We set `parallelism` to 1 because `wget` can't be parallelized.

Note how this pipeline itself has no inputs.  A pipeline that doesn't have inputs can only be triggered manually:

```
$ pachctl run-pipeline wordcount_input
```

Now you should be able to see a job running:

```
$ pachctl list-job
```

## wordcount_map

This pipeline counts the number of occurrences of each word it encounters.  While this task can very well be accomplished in Bash, we will demonstrate how to use custom code in Pachyderm by using a [Go program](map.go).

First of all, we build a [Docker image](Dockerfile) that contains our Go program:

```
$ docker build -t wordcount-map .
```

Then, we simply refer to the image in our pipeline:

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/mapPipeline.json
```

As soon as you create this pipeline, it will start processing data from `wordcount_input`.  Note that we did not specify a `parallelism` for this pipeline.  In this case, Pachyderm will automatically scale your pipeline based on the number of nodes in the cluster.

To parallelize the pipeline, Pachyderm spins up N concurrently running containers where each container gets `1/N` of the data.  For each word a container encounters, it writes a file whose filename is the word, and whose content is the number of occurrences.  If multiple containers write the same file, the content is merged.  As an example, the file `morning` might look like this:

```
$ pachctl list-job -p wordcount_map  # use this command to find out the output commit ID
ID                                 OUTPUT                                           STARTED             DURATION            STATE
6600c71be4e8604f716ce1965895dc27   wordcount_map/5b0e5c8f345b4f8c9690c77333564687   46 hours ago        -                   success
$ pachctl get-file wordcount_map 5b0e5c8f345b4f8c9690c77333564687 morning  # your commit ID will be different
217
355
142
```

This shows that there were three containers that wrote to the file `morning`.

## wordcount_reduce

The final pipeline goes through every file and adds up the numbers in each file.  For this pipeline we can use a simple bash script:

```
find /pfs/wordcount_map -name '*' | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count /pfs/out/`basename $count`; done
```

Which we bake into [reducePipeline.json](./reducePipeline.json)

```
# We assume you're running this from the root of this repo:
$ pachctl create-pipeline -f doc/examples/word_count/reducePipeline.json
```

The final output might look like this:

```
$ pachctl get-file wordcount_reduce [commit ID] morning
714
```

To get a complete list of the words counted:

```
$ pachctl list-file wordcount_reduce [commit ID]
```


## Preliminary Benchmarks

We ran this pipeline on 7.5GB of data in July 2016, on a 3-node GCE cluster with 4 CPUs and 15GB of memory each.  The job completed in 9 hours.
