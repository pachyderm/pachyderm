Tracking Your Data Flow
=======================
Once you've got a few pipelines built and have data flowing through the system, it becomes incredibly important to track that flow so you can attribute results to specific input data. Let's use the :doc:`beginner_tutorial` Fruit Stand as an example. 

Here is our data flow:

.. code-block:: shell

	+========+      +-----------+      +========+      +-----------+      +========+        
	|  Repo: | ===> | Pipeline: | ===> |  Repo: | ===> | Pipeline: | ===> |  Repo: | 
	|  Data  | ===> |  Filter   | ===> | Filter | ===> |    Sum    | ===> |   Sum  |
	+========+      +-----------+      +========+      +-----------+      +========+

Every commit of new log lines that comes into Data creates coinciding output commits on both the Filter and Sum repos. Let's say we want to programatically read the value of the file "apples" in Sum repo after each input commit. If a new commit, ``commit4``, is made now, how do we know when ``pachctl get-file sum master apples`` will be showing us the value of ``commit4`` and not ``commit3`` or ``commit5``? 

To do this, we'll use a feature of Pachyderm called Provenance. We're actually only using one piece of Provenance called ``flush-commit``. ``flush-commit`` will let our process block on an input commit until all of the output results are ready to read. 

You can read about the advanced features of Provenance, such as data lineage, in our Advanced :doc:`../advanced/provenance` Guide.

Using Flush-Commit
------------------

Synopsis:
Wait for all commits caused by the specified commit(s) to finish and return them.

Examples:

.. code-block:: shell

	# return all commits caused by Data/commit4
	$ pachctl flush-commit Data/commit4

	# return commits caused by Data/commit4 only in repo Sum
	$ pachctl flush-commit Data/Commmit4 -r Sum
	./pachctl flush-commit commit [commit ...]

	Options

	  -r, --repos value   Wait only for commits leading to a specific set of repos (default [])

