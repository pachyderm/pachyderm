Beginner Tutorial
=================
Welcome to the beginner tutorial for Pachyderm. If you've already got Pachyderm installed, this guide should take about 15 minutes and you'll be introduced to the basic concepts of Pachyderm.

Analyzing Log Lines from a Fruit Stand
--------------------------------------

In this guide you're going to create a Pachyderm pipeline to process
transaction logs from a fruit stand. We'll use two standard unix tools, ``grep``
and ``awk`` to do our processing. Thanks to Pachyderm's processing system we'll
be able to run the pipeline in a distributed, streaming fashion. As new data is
added, the pipeline will automatically process it and materialize the results.

If you hit any errors not covered in this guide, check our :doc:`troubleshooting` docs for common errors, submit an issue on `GitHub <https://github.com/pachyderm/pachyderm>`_, join our `users channel on Slack <http://slack.pachyderm.io>`_, or email us at `support@pachyderm.io <mailto:support@pachyderm.io>`_ and we can help you right away.

Prerequisites
^^^^^^^^^^^^^

This guide assumes that you already have Pachyderm running locally. Check out our :doc:`local_installation` instructions if haven't done that yet and then come back here to continue. 


Create a Repo
^^^^^^^^^^^^^

A ``repo`` is the highest level primitive in the Pachyderm file system (pfs). Like all primitives in pfs, it shares it's name with a primitive in Git and is designed to behave analogously. Generally, repos should be dedicated to a single source of data such as log messages from a particular service, a users table, or training data for an ML model. Repos are dirt cheap so don't be shy about making tons of them. 

For this demo, we'll simply create a repo called
"data" to hold the data we want to process:

.. code-block:: shell

 $ pachctl create-repo data

 # See the repo we just created
 $ pachctl list-repo
 data


Adding Data to Pachyderm
^^^^^^^^^^^^^^^^^^^^^^^^

Now that we've created a repo it's time to add some data. In Pachyderm, you write data to an explicit ``commit`` (again, similar to Git). Commits are immutable snapshots of your data which give Pachyderm its version control properties. ``Files`` can be added, removed, or updated in a given commit and then you can view a diff of those changes compared to a previous commit. 

Let's start by just adding a file to a new commit. We've provided a sample data file for you to use in our GitHub repo -- it's a list of purchases from a fruit stand. 

We'll use the ``put-file`` command along with two flags, ``-c`` and ``-f``. ``-f`` can take either a local file or a URL, in our case, the sample data on GitHub.

 We also specificy the repo name "data" and the branch name "master".

.. code-block:: shell

	$ pachctl put-file data master sales -c -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.2.0/doc/examples/fruit_stand/set1.txt

Unlike Git though, commits in Pachyderm must be explicitly started and finished as they can contain huge amounts of data and we don't want that much "dirty" data hanging around in an unpersisted state. The ``-c`` flag we used above specifies that we want to start a new commit, add data, and finish the commit in a convenient one-liner. 

Finally, we can see the data we just added to Pachyderm.

.. code-block:: shell

 # If we list the repos, we can see that there is now data
 $ pachctl list-repo
 NAME                CREATED             SIZE
 data                12 minutes ago       874 B

 # We can view the commit we just created
 pachctl list-commit data
 BRANCH              REPO/ID         PARENT              STARTED             FINISHED            SIZE
 master              data/master/0   <none>              6 minutes ago       6 minutes ago       874 B

 # We can also view the contents of the file that we just added
 $ pachctl get-file data master sales
 orange 	4
 banana 	2
 banana 	9
 orange 	9
 ...

Create a Pipeline
^^^^^^^^^^^^^^^^^

Now that we've got some data in our repo, it's time to do something with it.
``Pipelines`` are the core primitive for Pachyderm's processing system (pps) and
they're specified with a JSON encoding. For this example, we've already created the pipeline for you and it can be found at `examples/fruit_stand/pipeline.json on Github <https://github.com/pachyderm/pachyderm/blob/master/doc/examples/fruit_stand/pipeline.json>`_. Please open a new tab to view the pipeline while we talk through it.

When you want to create your own pipelines later, you can refer to the full :doc:`../development/pipeline_spec` to use more advanced options. This includes building your own code into a container instead of just using simple shell commands as we're doing here. 

For now, we're going to create a pipeline with 2 transformations in it. The first transformation filters the sales logs into separate records for apples,
oranges and bananas. The second step sums these sales numbers into a final sales count.

.. code-block:: shell

 +----------+     +--------------+     +------------+
 |input data| --> |filter pipline| --> |sum pipeline|
 +----------+     +--------------+     +------------+

In the first step of this pipeline, we are grepping for the terms "apple", "orange", and "banana" and writing that line to the corresponding file. Notice we read data from ``/pfs/data`` (``/pfs/[input_repo_name]``) and write data to ``/pfs/out/``. These are special local directories that Pachyderm creates within the container for you. All the input data will be found in ``/pfs/[input_repo_name]`` and your code should always write to ``/pfs/out``. 

The second step of this pipeline takes each file, removes the fruit name, and sums up the purchases. The output of our complete pipeline is three files, one for each type of fruit with a single number showing the total quantity sold. 

Now let's create the pipeline in Pachyderm:

.. code-block:: shell

 $ pachctl create-pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.2.0/doc/examples/fruit_stand/pipeline.json


What Happens When You Create a Pipeline
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Creating a pipeline tells Pachyderm to run your code on **every** finished
commit in a repo as well as **all future commits** that happen after the pipeline is created. Our repo already had a commit, so Pachyderm automatically
launched a ``job`` to process that data.

You can view the job with:

.. code-block:: shell

 $ pachctl list-job
	ID                                 OUTPUT                                       STARTED             DURATION             STATE
	90c74896fd227f319c3c19459aa7a22b   sum/e4060e15948c4b7b89947a02eace5dca/0       2 minutes ago       Less than a second   success
	67c30d70ba9d2179aa133255f5dc81db   filter/d737e9b7cfae40d4aa8a8871cdb9f783/0    3 minutes ago       2 seconds            success

Every pipeline creates a corresponding repo with the same name where it stores its output results. In our example, the "filter" transformation created a repo called "filter" which was the input to the "sum" transformation. The "sum" repo contains the final output files.

.. code-block:: shell

 $ pachctl list-repo
 NAME                CREATED             SIZE
 sum                 2 minutes ago       12 B
 filter              2 minutes ago       200 B
 data                19 minutes ago      874 B


Reading the Output
^^^^^^^^^^^^^^^^^^

 We can read the output data from the "sum" repo in the same fashion that we read the input data (except now we need to use an explicit commitID because the "sum" repo doesn't have a "master" branch:

.. code-block:: shell

 $ pachctl get-file sum e4060e15948c4b7b89947a02eace5dca/0 apple
 133


Processing More Data
^^^^^^^^^^^^^^^^^^^^

Pipelines will also automatically process the data from new commits as they are
created. Think of pipelines as being subscribed to any new commits that are
finished on their input repo(s). Also similar to Git, commits have a parental
structure that track how files change over time. In this case we're going to be adding more data to the same file "sales."

In our fruit stand example, this could be making a commit every hour with all the new purchases that happened in that timeframe. 

Let's create a new commit with our previous commit as the parent and add more sample data (set2.txt) to "sales":

.. code-block:: shell

  $ pachctl put-file data master sales -c -f https://raw.githubusercontent.com/pachyderm/pachyderm/v1.2.0/examples/fruit_stand/set2.txt

Adding a new commit of data will automatically trigger the pipeline to run on
the new data we've added. We'll see a corresponding commit to the output
"sum" repo with files "apple", "orange" and "banana" each containing the cumulative total of purchases. Let's read the "apples" file again and see the new total number of apples sold. 

.. code-block:: shell

 $ pachctl get-file sum 4092f4675650476ab0a3fde5b7780316/0 apple
 324

One thing that's interesting to note is that our pipeline is completely incremental. Since ``grep`` is a ``map`` operation, Pachyderm will only ``grep`` the new data from set2.txt instead of re-filtering all the data. If you look back at the "sum" pipeline, you'll notice the ``method`` and that our code uses ``/prev`` to compute the sum incrementally based upon our previous commit. You can learn more about incrementally in our advanced :doc:`../advanced/incrementality` docs.

We can view the parental structure of the commits we just created.

.. code-block:: shell

 $ pachctl list-commit data
 BRANCH              REPO/ID             PARENT              STARTED             FINISHED            SIZE
 master              data/master/0       <none>              19 minutes ago      19 minutes ago      874 B
 master              data/master/1       master/0            2 minutes ago       2 minutes ago       874 B


Exploring the File System
^^^^^^^^^^^^^^^^^^^^^^^^^
Another nifty feature of Pachyderm is that you can mount the file system locally to poke around and explore your data using FUSE. FUSE comes pre-installed on most Linux distributions. For OS X, you'll need to install `OSX FUSE <https://osxfuse.github.io/>`_.


The first thing we need to do is mount Pachyderm's filesystem (pfs).

First create the mount point:

.. code-block:: shell

    $ mkdir ~/pfs


And then mount it:

.. code-block:: bash

 # We background this process because it blocks.
 $ pachctl mount ~/pfs &


This will mount pfs on ``~/pfs`` you can inspect the filesystem like you would any
other local filesystem such as using ``ls`` or pointing your browser at it.

.. code-block:: shell

 # We can see our repos
 $ ls ~/pfs
 data   filter 	sum

 # And commits
 $ ls ~/pfs/sum
 4092f4675650476ab0a3fde5b7780316/1	4092f4675650476ab0a3fde5b7780316/0

.. note::

 Use ``pachctl unmount ~/pfs`` to unmount the filesystem. You can also use the ``-a`` flag to remove all Pachyderm FUSE mounts. 

Next Steps
^^^^^^^^^^
You've now got Pachyderm running locally with data and a pipeline! If you want to keep playing with Pachyderm locally, here are some ideas to expand on your working setup.

- Write a script to stream more data into Pachyderm. We already have one in Golang for you on `GitHub <https://github.com/pachyderm/pachyderm/tree/v1.2.0/doc/examples/fruit_stand/generate>`_ if you want to use it. 
- Add a new pipeline that does something interesting with the "sum" repo as an input.
- Add your own data set and ``grep`` for different terms. This example can be generalized to generic word count. 

You can also start learning some of the more advanced topics to develop analysis in Pachyderm:

- :doc:`../development/deploying_on_the_cloud`
- :doc:`../development/inputing_your_data` from other sources
- :doc:`../development/custom_pipelines` using your own code

We'd love to help and see what you come up with so submit any issues/questions you come across on `GitHub <https://github.com/pachyderm/pachyderm>`_ , `Slack <http://slack.pachyderm.io>`_ or email at dev@pachyderm.io if you want to show off anything nifty you've created! 