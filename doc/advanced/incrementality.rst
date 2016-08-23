# Incrementality

Incrementality is an advanced feature of Pachdyerm that let's you only process the "diff" of data for each run of a pipeline. Since Pachyderm's underlying storage is version controlled and diff-aware, your processing code can take advantage of that to maximize efficiency. 

Due to Math, incrementality only works one some types of computation.


## Partion and strategy

Incrementality is defined in the pipeline spec on a per-input basis. In other words for each input, you should specify whether you want only the new data (diff) exposed or if you want all past data available. 

For each pipeline input, you may specify a "method".  A method dictates exactly what happens in the pipeline when a commit comes into the input repo.

A method consists of two properties: partition unit and incrementality.

### Partition Unit
Partition unit specifies the granularity at which input data is parallelized across containers.  It can be of three values: 

* `block`: different blocks of the same file may be parelleized across containers.
* `file`: the files and/or directories residing under the root directory (/) must be grouped together.  For instance, if you have four files in a directory structure like: 

```
/foo 
/bar
/buzz
   /a
   /b
```
then there are only three top-level objects, `/foo`, `/bar`, and `/buzz`, each of which will remain grouped in the same container. 
* `repo`: the entire repo.  In this case, the input won't be partitioned at all. 

### Incrementality

Incrementality is a boolean flag that describes what data needs to be available when a new commit is made on an input repo. Namely, do you want to process only the new data in that commmit (the diff) or does all of the data need to be reprocessed?

For instance, if you have a repo with the file `/foo` in commit 1 and file `/bar` in commit 2, then:

* If the input is incremental, the first job sees file `/foo` and the second job sees file `/bar`.
* If the input is nonincremental, the first job sees file `/foo` and the second job sees file `/foo` and file `/bar`.

For convenience, we have defined aliases for the three most commonly used input methods: map, reduce, and global.  They are defined below:

|                | Block |  Top-level Objects |  Repo  |
|----------------|-------|--------------------|--------|
|   Incremental  |  map  |                    |        |
| Nonincremental |       |       reduce       | global |

If no method is specified, the `map` method (Block + Incremental) is used by default.


Example (Sum)
-------------

Sum is a great starting example for how to do processing incrementally. If your input is a list of values that is constantly having new lines appended and your output is the sum, using the previous run's results is a lot more efficient than recomputing every value every time.

TODO




