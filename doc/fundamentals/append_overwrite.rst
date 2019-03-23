Appending vs Overwriting Files
==============================

Introduction
------------

Pachyderm is designed to work with pipelined data processing in a
containerized environment. The Pachyderm File System (pfs) is a
file-based system that is distributed and supports data of all types of
files (binary, csv, json, images, etc) from many sources and users. That
data is processed in parallel across many different jobs and pipelines
using the Pachyderm Pipeline System (pps). The Pachyderm File System
(pfs) and Pachyderm Pipeline System (pps) are designed to work together
to get the right version of the right data to the right container at the
right time.

Among the many complexities you must consider are these:

-  Files can be put into pfs in “append” or “overwrite” mode.
-  `Pipeline definitions <../reference/pipeline_spec.html>`__ use `glob
   patterns <../reference/pipeline_spec.html#the-input-glob-pattern>`__
   to filter a view of `input
   repositories <../getting_data_into_pachyderm.html#data-repositories>`__.
-  The Pachyderm Pipeline System must merge data from what may be
   multiple containers running the same pipeline code, at the same time.

When you add in the ability to do “cross” and “union” operators on
multiple input repositories to those three considerations, it can be a
little confusing to understand what’s actually happening with your
files!

This document will take you through some of the advanced details, best
practices, and conventions for the following. If you’re unfamiliar with
the topic, each link below will take you to the basics.

-  `loading data into Pachyderm <../reference/pipeline_spec.html>`__,
-  using `glob
   patterns <../reference/pipeline_spec.html#the-input-glob-pattern>`__
   to filter the data to your pipelines, and
-  `processing data with your
   pipelines <../fundamentals/creating_analysis_pipelines.html>`__ and
   placing it in output repos.

Loading data into Pachyderm
---------------------------

Appending to files
~~~~~~~~~~~~~~~~~~

When putting files into a pfs repo via Pachyderm’s ``pachctl`` utility
or via the Pachyderm APIs, it’s vital to know about the default
behaviors of the ``put-file`` command. The following commands create the
repo “voterData” and place a local file called “OHVoterData.csv” into
it.

::

   $ pachctl create-repo voterData
   $ pachctl put-file voterData master -f OHVoterData.csv

The file will, by default, be placed into the top-level directory of
voterData with the name “OHVoterData.csv”. If the file were 153.8KiB,
running the command to list files in that repo would result in

::

   $ pachctl list-file voterData master
   COMMIT                           NAME             TYPE COMMITTED    SIZE
   8560235e7d854eae80aa03a33f8927eb /OHVoterData.csv file 1 second ago 153.8KiB

If you were to re-run the ``put-file`` command above, by default, the
file would be appended to itself and listing the repo would look like
this:

::

   $ pachctl list-file voterData master
   COMMIT                           NAME             TYPE COMMITTED     SIZE
   105aab526f064b58a351fe0783686c54 /OHVoterData.csv file 2 seconds ago 307.6KiB

In this case, any pipelines that use this repo for input will see an
updated file that has double the data in it. It may also have an
intermediate header row. (See `Specifying
Header/Footer <../cookbook/splitting.html?highlight=header#specifying-header-footer>`__
for details on headers and footers in files.) This is Pachyderm’s
default behavior. What if you want to overwrite the files?

Overwriting files
~~~~~~~~~~~~~~~~~

This is where the ``-o`` (or ``--overwrite``) flag comes in handy. It
will, as you’ve probably guessed, overwrite the file, rather than append
it.

::

   $ pachctl put-file voterData master -f OHVoterData.csv --overwrite
   $ pachctl list-file voterData master
   COMMIT                           NAME             TYPE COMMITTED    SIZE
   8560235e7d854eae80aa03a33f8927eb /OHVoterData.csv file 1 second ago 153.8KiB
   $ pachctl put-file voterData master -f OHVoterData.csv --overwrite
   $ pachctl list-file voterData master
   COMMIT                           NAME             TYPE COMMITTED    SIZE
   4876f99951cc4ea9929a6a213554ced8 /OHVoterData.csv file 1 second ago 153.8KiB

Deduplication
~~~~~~~~~~~~~

Pachyderm will deduplicate data loaded into input repositories. If you
were to load another file that hashed identically to “OHVoterData.csv”,
there would be one copy of the data in Pachyderm with two files’
metadata pointing to it. This works even when using the append behavior
above. If you were to put a file named OHVoterData2004.csv that was
identical to that first put-file of OHVoterData.csv, and then update
OHVoterData.csv as shown above, there would be two sets of bits in
Pachyderm:

-  a set of bits that would be returned when asking for the old branch
   of OHVoterData.csv & OHVoterData2004.csv and
-  a set of bits that would be appended to that first set to assemble
   the new, master branch of OHVoterData.csv.

This deduping happens with each file as it’s added to Pachyderm. We
don’t do deduping within the file (“intrafile deduplication”) because
this system is built to work with any file type.

.. important::  Pachyderm is smart about keeping the minimum set of bits in the object store and
                assembling the version of the file you (or your code!) have asked for.

Use cases for large datasets in single files
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Splitting large files in Pachyderm
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Unlike a system like Git,
which expects almost all of your files to be text,
Pachyderm does not do intra-file diffing
because we work with any file type:

* text
* JSON
* images
* video
* binary
* and so on…

Pachyderm diffs content at the per-file level.
Therefore,
if one bit in content of a file changes,
Pachyderm sees that as a new file.
Similarly,
Pachyderm can only distribute computation at the level of a single file;
if your data is only one large file,
it can only be processed by a single worker.

Because of these reasons, it’s pretty common to break up large files
into smaller chunks. For simple data types, Pachyderm provides the
``--split`` flag to ``put-file`` to automatically do this for you. For
more complex splitting patterns (e.g. ``avro`` or other binary formats),
you’ll need to manually split your data either at ingest or with a
Pachyderm pipeline.

Split and target-file flags
^^^^^^^^^^^^^^^^^^^^^^^^^^^

For common file types that are often used in data science, such as CSV,
line-delimited text files, JavaScript Object Notation (json) files,
Pachyderm includes the powerful ``--split``, ``--target-file-bytes`` and
``--target-file-datums`` flags.

``--split`` will divide those files into chunks based on what a “record”
is. In line-delimited files, it’s a line. In json files, it’s an object.
``--split`` takes one argument: ``line``, ``json`` or
``sql``.

.. note:: See the `Splitting Data for Distributed Processing <../cookbook/splitting.html#pg-dump-sql-support>`__  cookbook for more details on SQL support.

This argument tells Pachyderm how you want the file split into chunks. For
example, if you use ``--split line``, Pachyderm will only divide your
file on newline boundaries, never in the middle of a line. Along with
the ``--split`` flag, it’s common to use additional “target” flags to
get better control over the details of the split.

.. note:: We’ll call each of the chunks a “split-file” in this document.

-  ``--target-file-bytes`` will fill each of the split-files with data
   up to the number of bytes you specify, splitting on the nearest
   record boundary. Let’s say you have a line-delimited file of 50
   lines, with each line having about 20 bytes. If you use the flags
   ``--split lines --target-file-bytes 100``, you’ll see the input file
   split into about 10 files or so, each of which will have 5 or so
   lines. Each split-file’s size will hover above the target value of
   100 bytes, not going below 100 bytes until the last split-file, which
   may be less than 100 bytes.

-  ``--target-file-datums`` will attempt to fill each split-file with
   the number of datums you specify. Going back to that same
   line-delimited 50-line file above, if you use
   ``--split lines --target-file-datums 2``, you’ll see the file split
   into 50 split-files, each of which will have 2 lines.

-  Specifying both flags, ``--target-file-datums`` and
   ``--target-file-bytes``, will result in each split-file containing just
   enough data to satisfy whichever constraint is hit first. Pachyderm
   will split the file and then fill the first target split-file with
   line-based records until it hits the record limit. If it passes the
   target byte number with just one record, it will move on to the next
   split-file. If it hits the target datum number after adding another
   line, it will move on to the next split-file. Using the example
   above, if the flags supplied to put-file are
   ``--split lines --target-file-datums 2 --target-file-bytes 100``, it
   will have the same result as ``--target-file-datums 2``, since that’s
   the most compact constraint, and file sizes will hover around 40
   bytes.

What split data looks like in a Pachyderm repository
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Going back to our 50-line file example, let’s say that file is named
“my-data.txt”. We’ll create a repo named “line-data” and load
my-data.txt into Pachyderm with the following commands:

::
   
   $ pachctl create-repo line-data
   $ pachctl put-file line-data master -f my-data.txt --split line

After put-file is complete, list the files in the repo.

::
   
   $ pachctl list-file line-data master
   COMMIT                           NAME         TYPE COMMITTED          SIZE
   8cce4de3571f46459cbe4d7fe222a466 /my-data.txt dir  About a minute ago 1.071KiB

.. important:: The ``list-file`` command indicates that the line-oriented file we uploaded, “my-data.txt”, is actually a directory.

This file *looks* like a directory because
the ``--split`` flag has instructed Pachyderm to split the file up, and
it has created a directory with all the chunks in it. And, as you can
see below, each chunk will be put into a file. Those are the
split-files. Each split-file will be given a 16-character filename,
left-padded with 0.

.. note:: ``--split`` does not currently allow you to define more sophisticated file names.
         This is a set of features we'll add in future releases.
         (See `Issue 3568 <https://github.com/pachyderm/pachyderm/issues/3568>`__, for example).

Each filename will be numbered sequentially in hexadecimal. We
modify the command to list the contents of “my-data.txt”, and the output
reveals the naming structure used:

::

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

Appending to files with –split
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Combining ``--split`` with the default “append” behavior of
``pachctl put-file`` allows flexible and scalable processing of
record-oriented file data from external, legacy systems. Each of the
split-files will be deduplicated. You would have to ensure that
``put-file`` commands always have the ``--split`` flag.

``pachctl`` will reject the command if ``--split`` is not specified to
append a file that it was previously specified with an error like this

::
   
   could not put file at "/my-data.txt"; a file of type directory is already there

Pachyderm will ensure that only the added data will get reprocessed when
you append to a file using ``--split``. Each of the split-files is
subject to deduplication, so storage will be optimized. A large file
with many duplicate lines (or objects that hash identically) which you
with ``--split`` may actually take up less space in pfs than it does as
a single file outside of pfs.

Appending files can make for efficient processing in downstream
pipelines. For example, let’s say you have a file named “count.txt”
consisting of 5 lines

::
   
   One
   Two
   Three
   Four
   Five

Loading that local file into Pachyderm using ``--split`` with a command
like

::
   
   pachctl put-file line-data master count.txt -f ./count.txt --split line

will result in five files in a directory named “count.txt” in the input
repo, each of which will have the following contents

::
   
   count.txt/0000000000000000: One
   count.txt/0000000000000001: Two
   count.txt/0000000000000002: Three
   count.txt/0000000000000003: Four
   count.txt/0000000000000004: Five

This would result in five datums being processed in any pipelines that
use this repo.

Now, take a one-line file containing

::
   
   Six

and load it into Pachyderm appending it to the count.txt file. If that
file were named, “more-count.txt”, the command might look like

::

   pachctl put-file line-data master my-data.txt -f more-count.txt --split line

That will result in six files in the directory named “count.txt” in the
input repo, each of which will have the following contents

::

   count.txt/0000000000000000: One
   count.txt/0000000000000001: Two
   count.txt/0000000000000002: Three
   count.txt/0000000000000003: Four
   count.txt/0000000000000004: Five
   count.txt/0000000000000005: Six

This would result in one datum being processed in any pipelines that use
this repo: the new file ``count.txt/0000000000000005``.

Overwriting files with –split
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The behavior of Pachyderm when a file loaded with ``--split`` is
overwritten is simple to explain but subtle in its implications.
Remember that the loaded file will be split into those
sequentially-named files, as shown above. If any of those resulting
split-files hashes differently than the one it’s replacing, that will
cause the Pachyderm Pipeline System to process that data.

This can have important consequences for downstream processing. For
example, let’s say you have that same file named “count.txt” consisting
of 5 lines that we used in the previous example

::

   One
   Two
   Three
   Four
   Five

As discussed prior, loading that file into Pachyderm using ``--split``
will result in five files in a directory named “count.txt” in the input
repo, each of which will have the following contents

::

   count.txt/0000000000000000: One
   count.txt/0000000000000001: Two
   count.txt/0000000000000002: Three
   count.txt/0000000000000003: Four
   count.txt/0000000000000004: Five

This would result in five datums being processed in any pipelines that
use this repo.

Now, modify that file by inserting the word “Zero” on the first line.

::

   Zero
   One
   Two
   Three
   Four
   Five

Let’s upload it to Pachyderm using ``--split`` and ``--overwrite``.

::

   pachctl put-file line-data master count.txt -f ./count.txt --split line --overwrite

The input repo will now look like this

::

   count.txt/0000000000000000: Zero
   count.txt/0000000000000001: One
   count.txt/0000000000000002: Two
   count.txt/0000000000000003: Three
   count.txt/0000000000000004: Four
   count.txt/0000000000000005: Five

As far as Pachyderm is concerned,
every single file existing has changed,
and a new file has been added.
This is because the filename is taken into account when hashing the data for the pipeline.
While only one new piece of content is being stored,
``Zero``,
all six datums would be processed by a downstream pipeline.

It’s important to remember that what looks like a simple upsert can be a
kind of a `fencepost
error <https://en.wikipedia.org/wiki/Off-by-one_error#Fencepost_error>`__.
Being “off by one line” in your data can be expensive, consuming
processing resources you didn’t intend to spend.

Datums in Pachyderm pipelines
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The “datum” is the fundamental unit of data processing in Pachyderm
pipelines. It is defined at the file level and filtered by the “globs”
you specify in your pipelines. *What* makes a datum is defined by you.
How do you do that?

When creating a pipeline, you can specify one or more input repos.
Each of these will contain files.
Those files are filtered by the "glob" you specify in the pipeline's definition,
along with the input operators you use.
That determines how the datums you want your pipeline to process appear in the pipeline:
globs and input operators,
along with other pipeline configuration operators,
specify how you would like those datums orchestrated across your processing containers.
Pachyderm Pipeline System (pps) processes each datum individually in containers in pods,
using Pachyderm File System (pfs) to get the right data to the right code at the right time and merge the results.

To summarize:

-  **repos** contain *files* in pfs
-  **pipelines** filter and organize those files into *datums* for
   processing through *globs* and *input repo operators*
-  pps will use available resources to process each datum, using pfs to
   assign datums to containers and merge results in the pipeline’s
   output repo.

Let’s start with one of the simplest pipelines. The pipeline has a
single input repo, ``my-data``. All it does is copy data from its input
to its output.

::

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

With this configuration, the ``my-pipeline`` repo will always be a copy
of the ``my-data`` repo. Where it gets interesting is in the view of
jobs processed. Let’s say you have two data files and
you use the ``put-file`` command to load both of those into my-data

::

   $ pachctl put-file my-data master -f my-data-file-1.txt -f my-data-file-2.txt

Listing jobs will show that the job had 2 input datums, something like
this:

::
   
   $ pachctl list-job
   ID                               PIPELINE    STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE
   0517ff33742a4fada32d8d43d7adb108 my-pipeline 20 seconds ago Less than a second 0       2 + 0 / 2 3.218KiB 3.218KiB success

What if you had defined the pipeline to use the “/” glob, instead? That
``list-job`` output would’ve showed one datum, because it treats the
entire input directory as one datum.

::
   
   $ pachctl list-job
   ID                               PIPELINE    STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE
   aa436dbb53ba4cee9baaf84a1cc6717a my-pipeline 19 seconds ago Less than a second 0       1 + 0 / 1 3.218KiB 3.218KiB success

If we had written that pipeline to have a ``parallelism_spec`` of
greater than 1, there would have still been only one pod used to process
that data. You can find more detailed information on how to use
Pachyderm Pipeline System and globs to do sophisticated configurations
in the `Distributed
Computing <http://docs.pachyderm.io/en/latest/fundamentals/distributed_computing.html>`__
section of our documentation.

When you have loaded data via a ``--split`` flag, as discussed above,
you can use the glob to select the split-files to be sent to a pipeline.
A detailed discussion of this is available in the Pachyderm cookbook
section `Splitting Data for Distributed
Processing <http://docs.pachyderm.io/en/latest/cookbook/splitting.html#splitting-data-for-distributed-processing>`__.

Summary
~~~~~~~

Pachyderm provides powerful operators for combining and merging your
data through input operations and the glob operator. Each of these have
subtleties that are worth working through with concrete examples.
