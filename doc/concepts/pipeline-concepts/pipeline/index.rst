Pipeline
========

A pipeline is a Pachyderm primitive that is responsible for collecting data
from a specified source, or ``pfs/input``, transform it according to the pipeline
configuration file, and place it to the ``pfs/out`` directory in the ``pachd``
container. Each pipeline subscribes to a branch in the
input repository or repositories, gets triggered by a commit,
and runs your code to completion.

A minimum pipeline specification must include the following parameters:

- ``name`` — The name of your data pipeline. You can set an arbitrary
  name that is meaningful to the code you want to run to process the
  data in the specified repository.
- ``input`` — A location of the data that you want to process, such as a
  Pachyderm repository. You can specify multiple input
  repositories, as well as combine the repositories as union or cross
  pipelines.
  Depending on the anticipated results, you can unite and cross
  pipelines as needed in this field. For more information, see `Cross
  and Union <cross-union.html>`__.

  One very important property that is defined in the ``input`` field
  is the ``glob`` pattern that defines how Pachyderm breaks the data into
  individual processing units, called Datums. For more information, see
  `Datum <../datum/index.html>`__.

- ``transform`` — Specifies the code that you want to run against your
  data, such as a Python script. Also, specifies a Docker image that
  you want to use to run that script.

**Example:**

.. code:: bash

   {
     "pipeline": {
       "name": "wordcount"
     },
     "transform": {
       "image": "wordcount-image",
       "cmd": ["python3", "/my_python_code.py"]
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
