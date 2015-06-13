#Word Count

Word count is the hello world of distributed computing.
The goal is to count the number of times that words occur in a corpus of text.
For example the sentence: "Able was I ere I saw Elba." would have counts:
Able: 1
was: 1
I: 2
ere: 1
saw: 1
Elba: 1

This example assumes that you've already got a working Pachyderm cluster and
the data your want to analyze is stored at: pfs://data.

Word count is simple enough that we can implement it entirely using shell
commands and a stock ubuntu image, no need to install anything extra:

```shell
image ubuntu

input data

run mkdir -p /out/counts
run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
```

Let's walk through the example line by line to make sure we understand it.

`image ubuntu` run all of these commands in the ubuntu Docker image.

`input data` Make the directory `pfs://data` available inside containers as `/in/data`

`run mkdir -p /out/counts` Create a place in the `/out` directory for the pipelines to write the counts to.

`run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done`
This is the first interesting line of the Pachfile. It uses a few shell
commands to count the words in our data set and then record the counts to disk.
At the end of this step we'll have a file for each word. Keeping with our
example from before, we would have a file `/counts/Elba` and the content of
that file would be `1`.
These files are Pachyderm's equivalent of Hadoops key value pairs that are
emitted from a Map step.

`shuffle counts`
Pachyderm automatically parallelizes commands so they'll run faster. Now that
we've counted the occurences we need to get the counts for each word on to a
single machine so they can be added up. Suppose we had a cluster with 3 shards
which saw 1, 2 and 3 occurences of the word `foo` respectively. Each shard will
have a file `/counts/foo`. When we call shuffle those files will be
concatenated on to a single shard, so we'll have a file `/counts/foo` whose content is:

```
1
2
3
```

`run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done`
Lastly we need to sum up the values in these files, this line would turn /counts/foo `6`.

# Installing the pipeline:
To install the pipeline you do:
```shell
curl -XPOST pfs/pipeline/wordcount -T pachfile
curl -XPOST pfs/commit?commit=my_commit
```
This will kick off the pipeline on the data in my_commit
Results will become available at:
```
curl pfs/pipeline/wordcount/file/counts/foo?commit=my_commit
```
