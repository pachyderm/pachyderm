# Appending vs Overwriting Files

## Introduction

Pachyderm is designed to work with pipelined data processing in a
containerized environment.  The Pachyderm File System (pfs) is a
file-based system that is distributed and supports data of all types
of files (binary, csv, json, images, etc) from many sources and
users. That data is processed in parallel across many different jobs
and pipelines using the Pachyderm Pipeline System (pps). The Pachyderm
File System (pfs) and Pachyderm Pipeline System (pps) are designed to
work together to get the right version of the right data to the right
container at the right time.

Among the many complexities you must consider are these:

- Files can be put into pfs in "append" or "overwrite" mode.
- Pipeline definitions use "glob" operators to filter a view of input repositories.
- Output repositories must merge data from what may be multiple containers running the same pipeline code, at the same time.

When you add in the ability to do "cross" and "union" operators on
multiple input repositories to those three considerations, it can be a
little confusing to understand what's actually happening with your
files!

This document will take you through loading data into pachyderm,
consuming that data in input repos and placing data into output
repositories to show you the various options and conventions of
working with data in Pachyderm.

## Loading data into Pachyderm

### Overwriting files

When putting files into a pfs repo via Pachyderm's `pachctl` utility or via the Pachyderm APIs,
it's vital to know about the default behaviors of the `put-file`
command. The following commands create the repo "voterData" and place
a local file called "OHVoterData.csv" into it.
```
$ pachctl create-repo voterData
$ pachctl put-file voterData master -f OHVoterData.csv
```
The file will, by default, be placed into the top-level folder of
voterData with the name "OHVoterData.csv".  If the file were 153.8KiB,
running the command to list files in that repo would result in

```
$ pachctl list-file voterData master
COMMIT                           NAME             TYPE COMMITTED    SIZE
8560235e7d854eae80aa03a33f8927eb /OHVoterData.csv file 1 second ago 153.8KiB
```
If you were to re-run the `put-file` command, above, by default, the
file would be appended to itself and listing the repo would look like
this:
```
$ pachctl list-file voterData master
COMMIT                           NAME             TYPE COMMITTED     SIZE
105aab526f064b58a351fe0783686c54 /OHVoterData.csv file 2 seconds ago 307.6KiB
```

In this case, any pipelines that use this repo for input will see an updated file that has double the data in it.

### Appending to files

This is where the `-o` (or `--overwrite`) flag comes in handy.  It will, as you've probably guessed, overwrite the file, rather than append it.
```
$ pachctl put-file voterData master -f OHVoterData.csv --overwrite
$ pachctl list-file voterData master
COMMIT                           NAME             TYPE COMMITTED    SIZE
8560235e7d854eae80aa03a33f8927eb /OHVoterData.csv file 1 second ago 153.8KiB
$ pachctl put-file voterData master -f OHVoterData.csv --overwrite
$ pachctl list-file voterData master
COMMIT                           NAME             TYPE COMMITTED    SIZE
4876f99951cc4ea9929a6a213554ced8 /OHVoterData.csv file 1 second ago 153.8KiB
```
### Deduplication
Pachyderm will deduplicate data loaded into input repositories.  If you were to load another file that hashed identically to "OHVoterData.csv", there would be one copy of the data in Pachyderm with two files' metadata pointing to it.  This works even when using the append behavior above.  If you were to put a file named OHVoterData2004.csv that was identical to that first put-file of OHVoterData.csv, and then update OHVoterData.csv as shown above, there would be two sets of bits in Pachyderm:

- a set of bits that would be returned when asking for the old branch of OHVoterData.csv & OHVoterData2004.csv and
- a set of bits that would be appended to that first set to assemble the new, master branch of OHVoterData.csv.

In short, Pachyderm is smart about keeping the minimum set of bits in the object store and assembling the version of the file you (or your code!) have asked for.

### Use cases for large datasets in single files

#### Split and target-file flags
This default "append" behavior is useful for file types that are often used in data science: line-delimited files and JavaScript Object Notation (json) files. `pachctl` includes the powerful `--split`, `--target-file-bytes` and `--target-file-datums` flags to deal with those files.

- `--split`  will divide those files into chunks based on what a "record" is. In line-delimited files, it's a line.  In json files, it's an object. `--split` takes one argument: `line` or `json`.  Those chunks are themselves files in Pachyderm.  We'll call each of those chunks a "split-file" in this document.

- `--target-file-bytes` will fill each of the split-files with data up to the number of bytes you specify, splitting on the nearest record boundary.  Let's say you have a line-delimited file of 50 lines, with each line having about 20 bytes.  If you use the flags `--split lines --target-file-bytes 100`, you'll see the input file split into about 10 files or so, each of which will have 5 or so lines.  Each split-file's size will hover above the target value of 100 bytes, not going below 100 bytes until the last split-file, which  may be less than 100 bytes.

- `--target-file-datums` will attempt to fill each split-file with the number of datums you specify.  Going back to that same line-delimited 50-line file above, if you use `--split lines --target-file-datums 2`, you'll see the file split into 50 split-files, each of which will have 2 lines.

- Specifying both flags, `--target-file-datums` and `--target-file-bytes`, will result each split-file containing just enough data to satisfy whichever constraint is hit first.  Pachyderm will split the file and then fill the first target split-file with line-based records until it hits the record limit. If it passes the target byte number with just one record, it will move on to the next split-file.  If it hits the target datum number after adding another line, it will move on to the next split-file. Using the example above, if the flags supplied to put-file are `--split lines --target-file-datums 2 --target-file-bytes 100`, it will have the same result as `--target-file-datums 2`, since that's the most compact constraint, and file sizes will hover around 40 bytes.

#### What split data looks like in a Pachyderm repository

Going back to our 50-line file example, let's say that file is named "my-data.txt".  We'll create a repo named "line-data" and load my-data.txt into Pachyderm with the following commands:

```
$ pachctl create-repo line-data

$ pachctl put-file line-data master -f my-data.txt --split line
```
After put-file is complete, list the files in the repo.
```
$ pachctl list-file line-data master
COMMIT                           NAME         TYPE COMMITTED          SIZE
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt dir  About a minute ago 1.071KiB
```
Note that the `list-file` command indicates that the line-oriented file we uploaded, "my-data.txt" is actually a directory. That's because the `--split` flag has instructed Pachyderm to split the file up, and it has created a directory with all the chunks in it.  And, as you can see below,  each chunk will be put into a file.  Those are the split-files.  Each split-file will be given a 16-character filename, left-padded with 0.  Each filename will be numbered sequentially in hexadecimal. We modify the command to list the contents of "my-data.txt", and the output reveals the naming structure used:
```
$ pachctl list-file line-data master my-data.txt
COMMIT                           NAME                          TYPE COMMITTED          SIZE
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000000 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000001 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000002 file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000003 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000004 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000005 file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000006 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000007 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000008 file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000009 file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000a file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000b file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000c file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000d file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000e file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000000f file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000010 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000011 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000012 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000013 file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000014 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000015 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000016 file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000017 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000018 file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000019 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001a file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001b file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001c file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001d file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001e file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000001f file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000020 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000021 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000022 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000023 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000024 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000025 file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000026 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000027 file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000028 file About a minute ago 24B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000029 file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002a file About a minute ago 23B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002b file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002c file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002d file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002e file About a minute ago 22B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/000000000000002f file About a minute ago 21B
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000030 file About a minute ago 22B
COMMIT                           NAME                          TYPE COMMITTED          SIZE
8cce4de3571f46459cbe4d7fe222a466 /my-data.txt/0000000000000031 file About a minute ago 22B
```
#### Appending to files with --split
Combining `--split` with the default "append" behavior of `pachctl put-file` allows flexible and scalable processing of record-oriented file data from external, legacy systems.  Each of the split-files will be deduplicated.  You would have to ensure that `put-file` commands always have the `--split flag`. 


`pachctl` will reject the command if `--split` is not specified to append a file that it was previously specified with an error like this
```
could not put file at "/my-data.txt"; a file of type directory is already there
```
Pachyderm will ensure that only the added data will get reprocessed when you append to a file using `--split`.  Each of the split-files is subject to deduplication, so storage will be optimized.  A large file with many duplicate lines (or objects that hash identically) which you with `--split` may actually take up less space in pfs than it does as a single file outside of pfs.

Appending files can make for efficient processing in downstream pipelines.  For example, let's say you have a file named "count.txt" consisting of 5 lines

```
One
Two
Three
Four
Five
```
Loading that local file into Pachyderm using `--split` with a command like
```
pachctl put-file line-data master count.txt -f ./count.txt --split line
```
will result in five files in a directory named "count.txt" in the input repo, each of which will have the following contents
```
count.txt/0000000000000000: One
count.txt/0000000000000001: Two
count.txt/0000000000000002: Three
count.txt/0000000000000003: Four
count.txt/0000000000000004: Five
```
This would result in five datums being processed in any pipelines that use this repo.

Now, take a one-line file containing
```
Six
```
and load it into Pachyderm appending it to the count.txt file.  If that file were named, "more-count.txt", the command might look like
```
pachctl put-file line-data master my-data.txt -f more-count.txt --split line
```
That will result in six files in the directory named "count.txt" in the input repo, each of which will have the following contents
```
count.txt/0000000000000000: One
count.txt/0000000000000001: Two
count.txt/0000000000000002: Three
count.txt/0000000000000003: Four
count.txt/0000000000000004: Five
count.txt/0000000000000004: Six
```
This would result in one datum being processed in any pipelines that use this repo: the new file `count.txt/0000000000000004`.

#### Overwriting files with --split
The behavior of Pachyderm when a file loaded with `--split` is overwritten is simple to explain but subtle in its implications.  Remember that the loaded file will be split into those sequentially-named files, as shown above.  If any of those resulting split-files hashes differently than the one it's replacing, that will cause the Pachyderm Pipeline System to process that data.

This can have important consequences for downstream processing.  For example, let's say you have that same file named "count.txt" consisting of 5 lines that we used in the previous example
```
One
Two
Three
Four
Five
```
As discussed prior, loading that file into Pachyderm using `--split` will result in five files in a directory named "count.txt" in the input repo, each of which will have the following contents
```
count.txt/0000000000000000: One
count.txt/0000000000000001: Two
count.txt/0000000000000002: Three
count.txt/0000000000000003: Four
count.txt/0000000000000004: Five
```
This would result in five datums being processed in any pipelines that use this repo.

Now, modify that file by inserting the word "Zero" on the first line.
```
Zero
One
Two
Three
Four
Five
```
Let's upload it to Pachyderm using `--split` and `--overwrite`.
```
pachctl put-file line-data master count.txt -f ./count.txt --split line --overwrite
```
The input repo will now look like this
```
count.txt/0000000000000000: Zero
count.txt/0000000000000001: One
count.txt/0000000000000002: Two
count.txt/0000000000000003: Three
count.txt/0000000000000004: Four
count.txt/0000000000000005: Five
``` 
As far as Pachyderm is concerned, every single file existing has changed, and a new file has been added.  Six datums would be processed by a downstream pipelines.

It's important to remember that what looks like a simple upsert can be a kind of a [fencepost error](https://en.wikipedia.org/wiki/Off-by-one_error#Fencepost_error).  Being "off by one line" in your data can be expensive, consuming processing resources you didn't intend to spend.

### Datums in Pachyderm pipelines

The "datum" is the fundamental unit of data processing in Pachyderm pipelines.  It is defined at the file level and filtered by the "globs" you specify in your pipelines.  *What* makes a datum is defined by you.  How do you do that?

When creating a pipeline, you can specify one or more input repos.  Each of these will contain files.  Those files are filtered by the "glob" you specify in the pipeline's defnition, along with the input operators you use.  That determines how the datums you want your pipeline to process appear in the pipeline: globs and input operators, along with other pipeline configuration operators, specify how you would like those datums orchestrated across your processing containers. Pachyderm Pipeline System (pps) processes each datum individually in containers in pods, using Pachyderm File System (pfs) to get the right data to the right code at the right time and merge the results. 

To summarize:

- **repos** contain _files_ in pfs
- **pipelines** filter and organize those files into _datums_ for processing through _globs_ and _input repo operators_
- pps will use available resources to process each datum, using pfs to assign datums to containers and merge results in the pipeline's output repo.

Let's start with one of the simplest pipelines.  The pipeline has a single input repo, `my-data`.  All it does is copy data from its input to its output.
```
{
  "pipeline": {
    "name": "my-pipeline"
  },
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "my-data"
    }
  },
  "transform": {
      "cmd": ["sh"],
      "stdin": ["/bin/cp -r /pfs/my-data/\* /pfs/out/"],
    "image": "ubuntu:14.04"
  }
}
```
With this configuration, the `my-pipeline` repo will always be a copy of the `my-data` repo.  Where it gets interesting is in the view of jobs processed.  Let's say you have two data files listed in the file "data-files.txt", and you use the `put-file` command to load both of those into my-data
```
$ cat data-files.txt
my-data-file-1.txt
my-data-file-2.txt
$ pachctl put-file my-data master -i data-files.txt
```
Listing jobs will show that the job had 2 input datums, something like this:
```
$ pachctl list-job
ID                               PIPELINE    STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE
0517ff33742a4fada32d8d43d7adb108 my-pipeline 20 seconds ago Less than a second 0       2 + 0 / 2 3.218KiB 3.218KiB success
```
What if you had defined the pipeline to use the "/" glob, instead?  That `list-job` output would've showed one datum, because it treats the entire input directory as one datum.
```
$ pachctl list-job
ID                               PIPELINE    STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE
aa436dbb53ba4cee9baaf84a1cc6717a my-pipeline 19 seconds ago Less than a second 0       1 + 0 / 1 3.218KiB 3.218KiB success
```
If we had written that pipeline to have a `parallelism_spec` of greater than 1, there would have still been only one pod used to process that data.  You can find more detailed information on how to use Pachyderm Pipeline System and globs to do sophisticated configurations in the [Distributed Computing](http://docs.pachyderm.io/en/latest/fundamentals/distributed_computing.html) section of our documentation.

When you have loaded data via a `--split` flag, as discussed above, you can use the glob to select the split-files to be sent to a pipeline.  A detailed discussion of this is available in the Pachyderm cookbook section [Splitting Data for Distributed Processing](http://docs.pachyderm.io/en/latest/cookbook/splitting.html#splitting-data-for-distributed-processing).

### Lifecycle of a datum

There are four basic rules to how Pachyderm will process your data in the pipelines you create.

- Pachyderm will split your input into individual datums _[as you specified](./reference/pipeline_spec.html#the-input-glob-pattern)_ in your pipeline spec
- each datum will be processed independently, using _[the parallelism you specified](../reference/pipeline_spec.html#parallelism-spec-optional)_ in your pipeline spec, with no guarantee of the order of processing
- Pachyderm will merge the output of each pod into each output file, with no guarantee of ordering
- the output files will look as if everything was done in one processing step

If one of your pipelines is written to take all its input, whatever it may be, process it, and put the output into one file, the final output will be one file.  The many datums in your pipeline's input would become one file at the end, with the processing of every datum reflected in that file.

If you write a pipeline to write to multiple files, each of those files may contained merged data that looks as if it were all processed in one step, even if you would have specified the pipeline to process each datum in parallel.

The Pachyderm Pipeline System works with the Pachyderm File System to make sure that files output by each pipeline are merged successfully at the end of every commit.

#### Cross and union inputs

When creating pipelines, you can use "union" and "cross" operations to combine inputs.

[Union input](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#union-input) will combine each of the datums in the input repos as one set of datums. The result is that the number of datums processed is the sum of all the datums in each repo.

<img alt="Union input animation: shows how each datum is comprised of each datum in every input to the union." src="../_images/union.gif" width="200px" height="200px"/>

For example, let's say you have two input repos, A and B.  Each of them contain three files with the same names: my-data-file-1.txt, my-data-file-2.txt, and my-data-file-3.txt, but different content.  If you were to cross them in a pipeline, the "input" object in the pipeline spec might look like this.
```
  "input": {
      "union": [
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
If each repo had those three files at the top, there would be six (6) datums overall, which is the sum of the number of input files.  You'd see the following datums, in a random order, in your pipeline as it ran though them.
```
/pfs/A/my-data-file-1.txt

/pfs/A/my-data-file-2.txt

/pfs/A/my-data-file-3.txt

/pfs/B/my-data-file-1.txt

/pfs/B/my-data-file-2.txt

/pfs/B/my-data-file-3.txt
```
One of the ways you can make your code simpler is to use the `name` field for the `pfs` object and give each of those repos the same name.
```
  "input": {
      "union": [
          {
              "pfs": {
                  "name": "C",
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "name": "C",
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
You'd still see the same six datums, in a random order, in your pipeline as it ran though them, but they'd all be in a directory with the same name: C.
```
/pfs/C/my-data-file-1.txt  # from A

/pfs/C/my-data-file-2.txt  # from A

/pfs/C/my-data-file-3.txt  # from A

/pfs/C/my-data-file-1.txt  # from B

/pfs/C/my-data-file-2.txt  # from B

/pfs/C/my-data-file-3.txt  # from B
```
[Cross input](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#cross-input) is essentially a cross-product of all the datums, selected by the globs on the repos you're crossing.  It provides a combination of all the datums to the pipeline that uses it as input, treating each combination as a datum.  

<img alt="Cross input animation: shows how each datum in the pipeline is actually one of the  combinations of every datum across the inputs to the cross" src="../_images/cross.gif" width="200px" height="200px">

There are many examples that show the power of this operator: [Combining/Merging/Joining Data](http://docs.pachyderm.io/en/latest/cookbook/combining.html#combining-merging-joining-data) cookbook and the [Distributed hyperparameter tuning](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter) example are good ones.  

It's important to remember that paths in your input repo will be preserved and prefixed by the repo name to prevent collisions between identically-named files.  For example, let's take those same two input repos, A and B, each of which with the same files as above.  If you were to cross them in a pipeline, the "input" object in the pipeline spec might look like this
```
  "input": {
      "cross": [
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
In the case of cross inputs, you can't give the repos being crossed identical names, because of that need to avoid name collisions. If each repo had those three files at the top, there would be nine (9) datums overall, which is every permutation of the input files.  You'd see the following datums, in a random order, in your pipeline as it ran though the nine permutations. 
```
/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-2.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-3.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-2.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-3.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-3.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-3.txt
```
Note that you always see both input directories involved in the cross!
### Output repositories

Every Pachyderm pipeline has an output repository associated with it, with the same name as the pipeline.  [For example](../getting_started/beginner_tutorial.html), an "edges" pipeline would have an "edges" repository you can use as input to other pipelines, like a "montage" pipeline.

Your code, regardless of the pipeline you put it in, should place data in a filesystem mounted under "/pfs/out" and it will appear in the named repository for the current pipeline.

### Appending vs overwriting data in output repositories

The Pachyderm File System keeps track of which datums are being processed in which containers, and makes sure that each datum leaves its unique data in output files.  Let's say you have a simple pipeline, "[wordcount](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count)", that counts references to words in documents by writing the number of occurrences of a word to an output file named for each word in `/pfs/out`, followed by a newline. We intend to process the data by treating each input file as a datum.  We specify the glob in the "wordcount" pipeline json accordingly, something like `"glob": "/*"`. We load a file containing the first paragraph of Charles Dickens's "A Tale of Two Cities" into our input repo, but mistakenly just put the first four lines in the file `tts.txt`.
```
It was the best of times,
it was the worst of times,
it was the age of wisdom,
it was the age of foolishness,
```
In this case, after the pipeline runs on this single datum, `/pfs/out` would contain the files
```
it -> 4\n
was -> 4\n
the -> 4\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
Where "\n" is the newline appended by our "wordcount" code after it outputs the word count. If we were to fix `tts.txt`, either by appending the missing text or replacing it with the entire first paragraph using `pachctl put-file` with the `--overwrite` flag, the file would then look like this
```
It was the best of times,
it was the worst of times,
it was the age of wisdom,
it was the age of foolishness,
it was the epoch of belief,
it was the epoch of incredulity,
it was the season of Light,
it was the season of Darkness,
it was the spring of hope,
it was the winter of despair,
we had everything before us,
we had nothing before us,
we were all going direct to Heaven,
we were all going direct the other way--
in short, the period was so far like the present period, that some of
its noisiest authorities insisted on its being received, for good or for
evil, in the superlative degree of comparison only.
```
We would see each file in the "wordcount" repo overwritten with one line with an updated number. Using our existing examples, we'd see a few of the files replaced with new content
```
it -> 10\n
was -> 10\n
the -> 14\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
The reason that the entire file gets reprocessed, even if we just append to it, is because the entire file is the datum.  We haven't used the `--split` flag combined with the appropriate glob to split it into lots of datums.

What if we have other texts in the pipeline, other datums?  Like the first paragraph of Herman Melville's Moby Dick, put into md.txt?
```
Call me Ishmael. Some years ago—never mind how long precisely—having
little or no money in my purse, and nothing particular to interest me
on shore, I thought I would sail about a little and see the watery part of the world. 
It is a way I have of driving off the spleen and
regulating the circulation. Whenever I find myself growing grim about
the mouth; whenever it is a damp, drizzly November in my soul; whenever
I find myself involuntarily pausing before coffin warehouses, and
bringing up the rear of every funeral I meet; and especially whenever
my hypos get such an upper hand of me, that it requires a strong moral
principle to prevent me from deliberately stepping into the street, and
methodically knocking people's hats off—then, I account it high time to
get to sea as soon as I can. This is my substitute for pistol and ball.
With a philosophical flourish Cato throws himself upon his sword; I
quietly take to the ship. There is nothing surprising in this. If they
but knew it , almost all men in their degree, some time or other,
cherish very nearly the same feelings towards the ocean with me.
```
What happens to our word files?  Will they get overwritten?  Not as long as each input file--`tts.txt` and `md.txt`--is being treated as a separate datum.  You'll see the data in the "wordcount" repo looking something like this:
```
it -> 10\n5\n
was -> 10\n
the -> 14\n7\n
best -> 1\n
worst -> 1\n
of -> 4\n4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
During each job that is run, each distinct datum in Pachyderm will put data in an output file.  If the file shares a name with the files from other datums,  they're merged after processing. This will happen during the appropriately-named _merge_ stage after your pipeline runs. You should not count on the data appearing in a particular order, or the data in a file in an output repo from a particular datum containing the data from the processing of any other datums.  You won't see it in the file until all datums have been processed and the merge is complete, after that pipeline is finished.

What happens if we delete `md.txt`?  The "wordcount" repo would go back to its condition with just `tts.txt`.
```
it -> 10\n
was -> 10\n
the -> 14\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
What if didn't delete `md.txt`; we appended to it?  Then we'd see the appropriate counts change only on the lines of the files affected by `md.txt`; the counts for `tts.txt` would not change. Let's say we append the second paragraph to `md.txt`:
```
There now is your insular city of the Manhattoes, belted round by
wharves as Indian isles by coral reefs—commerce surrounds it with her
surf. Right and left, the streets take you waterward. Its extreme
downtown is the battery, where that noble mole is washed by waves, and
cooled by breezes, which a few hours previous were out of sight of
land. Look at the crowds of water-gazers there.
```
The "wordcount" repo might now look like this. (We're not using stemmed parser, and "it" is a different word than "its")
```
it -> 10\n6\n
was -> 10\n
the -> 14\n11\n
best -> 1\n
worst -> 1\n
of -> 4\n8\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
Pachyderm is smart enough to keep track of what changes to what datums affect what downstream results, and only reprocesses and re-merges as needed.

To summarize, 

- If you overwrite an input datum in pachyderm with new data, each output datum that can trace to that input datum in downstream pipelines will get updated 
- If your downstream pipeline processes multiple input datums, putting the result a single file, adding or removing an input datum will only remove its effect from that file.  The effect of the other datums will still be seen in that file.

You can see this in action in the [word count example](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count) in the Pachyderm git repo.

## Deduplication of output data

Data is not currently deduplicated globally across output repos for performance reasons, only locally, within that repo. If you copy data from an output repo to an input repo, it will be deduplicated globally.

### Summary

Pachyderm provides powerful operators for combining and merging your data through input operations and the glob operator.  Each of these have subtleties that are worth working through with concrete examples.
