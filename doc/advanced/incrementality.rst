Incrementality
==============

Incrementality is an advanced feature of Pachdyerm that let's you only process the "diff" of data for each run of a pipeline. Since Pachyderm's underlying storage is version controlled and diff-aware, your processing code can take advantage of that to maximize efficiency. 

Due to Math, incrementality only works one some types of computation.

Method
------

Incrementality is defined in the pipeline spec on a per-input basis. In other words for each input, you should specify whether you want only the new data (diff) exposed or if you want all past data available. 

For each pipeline input, you may specify a "method".  A method dictates exactly what happens in the pipeline when a commit comes into the input repo.

A method consists of two properties: partition unit and incrementality.

Partition Unit
^^^^^^^^^^^^^^

Partition unit specifies the granularity at which input data is parallelized across containers.  It can be of three values: 

1. ``block``: different blocks of the same file may be parelleized across containers.

2. ``Top-level Objects``: the files and/or directories residing under the root directory (/) must be grouped together.  For instance, if you have four files in a directory structure like: 

.. code-block:: shell

	/foo 
	/bar
	/baz
	   /a
	   /b

then there are only three top-level objects, ``/foo``, ``/bar``, and ``/baz``. ``/baz/a`` and ``/baz/b`` will always be seen by the same container but there are no guarantees about where ``foo`` or ``bar`` are processed relative to ``baz``. 

3. ``repo``: the entire repo.  In this case, the input won't be partitioned at all and all data in the repo will be available. 


Incrementality
^^^^^^^^^^^^^^


Incrementality is a numerical flag (0, 1, or 2) that describes what data needs to be available when a new commit is made on an input repo. Namely, do you want to process only the new data in that commmit (the diff) or does all of the data need to be reprocessed?

For instance, if you have a repo with the file `/foo` in commit 1 and file `/bar` in commit 2, then:

* If the input is incremental (1), the first job sees file `/foo` and the second job sees file `/bar`.
* If the input is nonincremental(0), ever job sees all the data. The first job sees file `/foo` and the second job sees file `/foo` and file `/bar`.

Top-level objects (2) means that if any part in a file (or any file within a directory) changes, then show all the data in that file (directory). For example, you may have a directory called "users" with each user's info as a file. `Incremental: 2` would mean that if any user file changed, your job should see all user files as input.

For convenience, we have defined aliases for the three most commonly used input methods: map, reduce, and global.  They are defined below:


+---------------------+----------+-----------------------+----------+
|                     |  "Block" |  "File" (Top-lvl Obj) |  "Repo"  |
+=====================+==========+=======================+==========+
| 0 (non-incremental) |          |        reduce         |  global  |
+---------------------+----------+-----------------------+----------+
| 1 (diffs)           |    map   |                       |          |
+---------------------+----------+-----------------------+----------+
| 2 (top-lvl object)  |          |                       |          |
+---------------------+----------+-----------------------+----------+


If no method is specified, the ``map`` method (Block + Incremental) is used by default.

Writing Incremental Code
------------------------

Writing your analysis code to take advantage of incrementality involves understanding two ideas: What new data is available as input and what was the output last time this pipeline ran (aka: output parent commit)? To answer both of these questions, you need to understand which data is exposed where in Pachyderm. 

Mount Path
^^^^^^^^^^

The root mount point is at ``/pfs``, which contains:

- ``/pfs/input_repo`` which is where you would find the latest commit from each input repo you specified.
  - Each input repo will be found here by name
  - Note: Unlike when mounting pfs locally, there is no CommitID in the path. This is because the commit will always change, and the ID isn't relevant to the processing. The commit that is exposed is configured based on the incrementality flag above.
- ``/pfs/out`` which is where you write any output
- ``/pfs/prev`` which is this pipeline's previous output, if it exists. (You can think of it as this job's output commit's parent). 

The easiest way to understand how to use incrementality and ``/pfs/prev`` is through a simple example.

Example (Sum)
^^^^^^^^^^^^^

Sum is a great starting example for how to do processing incrementally. If your input is a list of values that is constantly having new lines appended and your output is the sum, using the previous run's results is a lot more efficient than recomputing every value every time.

First, we should set ``partition: 2`` ("file") and ``Incremental: true``. Setting partition in this way ensures that all the values are seen by one container. If we had this set to ``map`` instead, we may get some input values spread across containers and we wouldn't get an accurate total. Incremental ansures that only the new values are shown.

For each run of the pipeline, ``/pfs/<input_data>`` will be a file with all the new values that have been added in the most recent commit. Our pipeline should simply sum up those new values and add them to the previous total in ``/pfs/prev`` and write that new total to ``/pfs/out``. 




