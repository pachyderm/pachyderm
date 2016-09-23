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

1. ``BLOCK``: different blocks of the same file may be parelleized across containers.

2. ``FILE`` (Top-level Objects): the files and/or directories residing under the root directory (/) must be grouped together.  For instance, if you have four files in a directory structure like: 

.. code-block:: shell

	/foo 
	/bar
	/baz
	   /a
	   /b

then there are only three top-level objects, ``/foo``, ``/bar``, and ``/baz``. ``/baz/a`` and ``/baz/b`` will always be seen by the same container but there are no guarantees about where ``foo`` or ``bar`` are processed relative to ``baz``. 

3. ``REPO``: the entire repo.  In this case, the input won't be partitioned at all and all data in the repo will be available. 


Incrementality
^^^^^^^^^^^^^^

#### Incrementality

Incrementality ("NONE", "DIFF" or "FILE") describes what data needs to be available when a new commit is made on an input repo. Namely, do you want to process *only the new data* in that commmit (the "diff"), only files with any new data ("FILE"), or does all of the data need to be reprocessed ("NONE")?

For instance, if you have a repo with the file ``/foo`` in commit 1 and file ``/bar`` in commit 2, then:

* If the input incrementality is "DIFF", the first job sees file ``/foo`` and the second job sees file ``/bar``.

* If the input is non-incremental("NONE"), every job sees all the data. The first job sees file ``/foo`` and the second job sees file ``/foo`` *and* file ``/bar``.

* "FILE" (Top-level objects) means that if any part in a file (or alternatively any file within a directory) changes, then show all the data in that file (directory). For example, you may have vendor data files in separate directories by state -- the California directory contains a file for each california vendor, etc.  ``Incremental: "FILE"`` would mean that your job will see the entire directory if at least one file in that directory has changed. If only one vendor file in the whole repo was was changed and it was in the Colorado directory, all Colorado vendor files would be present, but that's it. 

#### Combining Partition unit and Incrementality

For convenience, we have defined aliases for the three most commonly used (and most familiar) input methods: "map", "reduce", and "global". 

* A "map" (BLOCK + DIFF), for example, can partition files at the block level and jobs only need to see the new data. 

* "Reduce" (FILE + NONE) as it's typically seen in Hadoop, requires all parts of a file to be seen by the same container ("FILE") and your job needs to reprocess *all* the data in the whole repo ("NONE"). 

* "Global" (REPO + NONE), means that the entire repo needs to be seen by *every* container. This is commonly used if you had a repo with just parameters, and every container needed to see all the parameter data and pull out the ones that are relevant to it. 

They are defined below:

.. code-block:: shell

                             +-----------------------------------------+
                             |             Partition Unit              |
  +--------------------------+---------+----------------------+--------+
  |     Incrementality       | "BLOCK" | "FILE" (Top-lvl Obj) | "REPO" |
  +==========================+=========+======================+========+
  | "NONE" (non-incremental) |         |       "reduce"       |"global"|
  +--------------------------+---------+----------------------+--------+
  | "DIFF" (incremental)     |  "map"  |                      |        |
  +--------------------------+---------+----------------------+--------+
  | "FILE" (top-lvl object)  |         |                      |        |
  +--------------------------+---------+----------------------+--------+

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

First, we should set ``partition: "FILE" and ``Incremental: "DIFF"``. Setting partition in this way ensures that all the values are seen by one container. If we had this set to ``map`` instead, we may get some input values spread across containers and we wouldn't get an accurate total. Incremental ansures that only the new values are shown.

For each run of the pipeline, ``/pfs/<input_data>`` will be a file with all the new values that have been added in the most recent commit. Our pipeline should simply sum up those new values and add them to the previous total in ``/pfs/prev`` and write that new total to ``/pfs/out``. 

