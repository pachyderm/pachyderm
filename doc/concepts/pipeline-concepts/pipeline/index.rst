Pipeline
========

A pipeline is a PPS primitive that is responsible for collecting data from
a specified source or ``pfs/input``, transform it according to the pipeline
configuration file, and place it to the ``pfs/out`` directory in your
Pachyderm repository. Each pipeline subscribes to a branch in the
input repository or repositories, gets triggered by a commit,
and runs your code to completion.

A minimum pipeline specification must include the following parameters:

- ``name`` - The name of your data pipeline. You can set an arbitrary
  name that is meaningful to the code you want to run to process the
  data in the specified repository.
- ``input`` - A location of data that you want to process, such as a
  Pachyderm repository or a Git repository. You can specify not only a
  repository, but also a combination, or a union, of multiple
  repositories, as well as cross-pipelines and cron pipelines.
  Depending on the anticipated results, you can unite and cross
  pipelines as needed in this field. For more information, see `Cross
  and Union <cross-union.html>`__.

  Another important property that you can specify in the ``input`` field
  is the ``glob`` pattern that defines how Pachyderm breaks the data into
  processing units. For more information, see
  `Datum <../datum/index.html>`__.

- ``transform`` - Specifies a code that you want to run against your
  data, such as a Python script and a Docker image that you want to use
  to run that script.

**Example:**

.. code:: bash

   {
     "pipeline": {
       "name": "wordcount"
     },
     "transform": {
       "image": "wordcount-image",
       "cmd": ["/binary", "/pfs/data", "/pfs/out"]
     },
     "input": {
           "pfs": {
               "repo": "data",
               "glob": "/*"
           }
       }
   }

.. toctree::
   :maxdepth: 2

   cross-union.md
   cron.md

**See also**

-  `Pipeline Specification <../../../reference/pipeline_spec.html>`__
