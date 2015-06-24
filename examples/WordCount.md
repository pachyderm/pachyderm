# Quickstart Guide
## Tutorial -- Distributed Word Count
__Requirements:__ This tutorial assumes you already have Docker v1.5 and btrfs tools v3.14 or higher installed.

Word count is the hello world of distributed computing. The goal is to count the number of times that words occur in a corpus of text. For example the sentence: "Able was I ere I saw Elba." would have counts:
Able: 1
was: 1
I: 2
ere: 1
saw: 1
Elba: 1

First, we're going to run the pipeline locally on a small dataset and then we'll seamlessly scale it up to a full cluster.
### Run the wordcount pipeline in Pachyderm locally
#### Step 1: Launch Pachyderm locally
Download and run the Pachyderm launch script to get a local instance running. It defaults to port 650.
```shell
# launch a local pfs instance
$ curl www.pachyderm.io/launch | sh
```
#### Step 2: Add a few text files to Pachyderm in the directory `text`
```shell
# add a local file to Pachyderm
$ curl localhost:650/file/text/textfile1 -T your_text_file
```
`file` is a Pachyderm keyword that lets us manipulate files in pfs. We're adding the text file in the directory /text and naming it `textfile1`. Read about the API in more detail [here](https://github.com/pachyderm/pfs/#the-pachyderm-http-api).

#### Step 3: Create the wordcount pipeline
In Pachyderm, pipelines are how we define distributed computations. The Docker image, input data, and analysis logic are defined in a Pachfile. We've already created the wordcount Pachfile for you, but we're going to go through it in detail to understand what's going on and how to create your own Pachfiles in the future. If you want to skip these details and just run the pipeline, [jump to step 4](https://github.com/pachyderm/pfs/blob/next/examples/WordCount.md#step-4-install-and-run-the-wordcount-pipeline-locally).

Word count is simple enough that we can implement it entirely using shell commands and a stock ubuntu image, no need to install anything extra. Here's the body for the wordcount Pachfile. It's only a few lines. [Download the wordcount Pachfile]() or copy/paste the text into your own text file.  

```shell
image ubuntu

input text

run mkdir -p /out/counts
run cat /in/text/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
```
Let's walk through the wordcount Pachfile line-by-line to make sure we understand it.

`image ubuntu`: run all of these commands in the ubuntu Docker image.

`input text`: Make the directory `text` in pfs available as `/in/text` to our analysis logic. This `in` directory is part of the magic of Pachyderm. All of your analysis logic will read data from `/in` and output to `/out`. Our analysis logic accesses the data at `/in` as a local file system which is why we can use whatever tools we want -- in this case shell commands.  

`run mkdir -p /out/counts` Creates the directory `/out/counts` for the pipelines to write to.

`run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done`:  This is the first  line of the Pachfile that makes up our analysis logic. It uses a few shell
commands to count the words in our data set and then records the counts to disk. At the end of this step we'll have a file for each word. Keeping with our example from before, we would have a file `/counts/Elba` and the content of
that file would be `1`. These files are Pachyderm's equivalent of Hadoops key-value pairs that are
emitted from a Map step.

`shuffle counts`: Pachyderm automatically parallelizes commands so they'll run faster. Now that
we've counted the occurences of each word, we need to get the counts for each word on to a
single machine so they can be added up. Suppose we had a cluster with 3 shards
which saw 1, 2 and 3 occurences of the word `foo`, respectively. Each shard will
have a file `/counts/foo`. When we call shuffle those files will be
concatenated onto a single shard, so we'll have a file `/counts/foo` whose content is:
```
1
2
3
```

`run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done`: Lastly we need to sum up the values in these files, this line would make /counts/foo have the content `6`.

#### Step 4: Install and run the wordcount pipeline locally
Assuming you've [downloaded the wordcount Pachfile]() (or created it yourself), we'll now POST it to the filesystem just like we did with the text files. Instead of using the keyword `file`, we use the keyword `pipeline` and we name this Pachfile `wordcount`
```shell
$ curl -XPOST localhost:650/pipeline/wordcount -T wordcount.pachfile
```
We've now added both our data and pipeline to Pachyderm and we want to `commit` both of them. `commit` is another Pachyderm keyword that creates an immutable snapshot of the data and Pachfiles. Creating a commit also runs all of the analysis pipelines in the system. In the example below, we've named our commit `commit1`.
```shell
$ curl -XPOST localhost:650/commit?commit=commit1
```
Results will become available at:
```
$ curl localhost:650/pipeline/wordcount/file/counts/Elba?commit=commit1
```
If you don't see any results, make sure the pipeline has finished running. Replace `Ebla` with whatever word you'd like to see the count for. If you've made multiple commits you can see the results of a past commit by changing the commit name. 

Next, let's deploy a full cluster and run the same exact pipeline on way more data! 

### Run the wordcount pipeline in a Pachyderm cluster
Some of these steps look really similar to those above -- that's the whole point! Running pipelines in a cluster _should_ be just as easy as running and testing them locally. 

#### Step 5: Deploy a Pachyderm Cluster
The easiest deployment option is to use the AWS cloud template we've built for you.
- [Deploy on Amazon EC2](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=Pachyderm&templateURL=https:%2F%2Fs3-us-west-1.amazonaws.com%2Fpachyderm-templates%2Ftemplate) using cloud templates (recommended)
 If you prefer to use a different host or set up your cluster manually, see [Cluster Deployment](https://github.com/pachyderm/pfs#creating-a-pachyderm-cluster)

#### Step 6: Add your text files to Pachyderm
This step is exactly the same as step 2 except we replace localhost:650 with the `<hostname>` of one of our EC2 machines. Use the hostname of any machine in the cluster and pachyderm will automatically distributed the files across all machines and shards.
```shell
# add a file to the Pachyderm cluster
$ curl <hostname>/file/data/textfile1 -T your_text_file
```
Go ahead and add a whole bunch of text files!

#### Step 7: Install and run the wordcount pipeline in the cluster
```shell
curl -XPOST <hostname>/pipeline/wordcount -T wordcount.pachfile
curl -XPOST <hostname>/commit?commit=commit1
```

And once again, results will become available at:
```
curl <hostname>/pipeline/wordcount/file/counts/Elba?commit=commit1
```
#### Step 8: Editing the wordcount Pachfile
If you want to do something slightly different than wordcount, it's really easy to change the analysis by editing the Pachfile. EXAMPLE: Only list > 2? top 5 words?

