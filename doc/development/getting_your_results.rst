Getting your Results
====================
Once you've got a few pipelines built and have data flowing through the system, it becomes incredibly important to track that flow so you can read the correct results. Let's use the :doc:`../getting_started/beginner_tutorial` Fruit Stand as an example. 

Here is our data flow:

.. code-block:: shell

	+========+      +-----------+      +========+      +-----------+      +========+        
	|  Repo: | ===> | Pipeline: | ===> |  Repo: | ===> | Pipeline: | ===> |  Repo: | 
	|  Data  | ===> |  Filter   | ===> | Filter | ===> |    Sum    | ===> |   Sum  |
	+========+      +-----------+      +========+      +-----------+      +========+

Every commit of new log lines that comes into Data creates corresponding output commits on both the Filter and Sum repos. Let's say we want to programatically read the value of the file "apples" in Sum repo after each input commit. If a new commit, ``master/4``, is made now, how do we know when ``pachctl get-file sum master apples`` will be showing us the resulting value of ``master/4`` and not ``master/3`` or ``master/5``? 

To do this, we'll use a feature of Pachyderm called Provenance. We're actually only using one piece of Provenance called ``flush-commit``. ``flush-commit`` will let our process block on an input commit until all of the output results are ready to read. In other words, ``flush-commit`` lets you view a consistent global snapshot of all your data at a given commit. 

You can read about other advanced features of Provenance, such as data lineage, in our Advanced :doc:`../advanced/provenance` Guide, but we're just going to cover ``flush-commit`` here. 


Using Flush-Commit
------------------
Let's demonstrate a typical workflow using ``flush-commit`` First, we'll make a new commit into the Data repo, ``master/4``. The filter and sum pipelines (in serial) are chugging along and we want to read out the result of "apples" after the new data in ``master/4`` has fully propogated through our pipelines. We do this with:

.. code-block:: shell

	# return the commit in Sum caused by Data/master/commit4 (<repo_name/commitID>)
	$ pachctl flush-commit Data/master/4
	BRANCH                             Repo/ID                                    PARENT      STARTED             FINISHED            SIZE
	master                             Data/master/4                              <none>      55 seconds ago   	  55 seconds ago      874 B
	fa59744b0e0448348159fef216f4eee9   Sum/fa59744b0e0448348159fef216f4eee9/0     <none>      37 seconds ago      36 seconds ago      12 B
	557c7db83002419aa2634e0c0ca9f2e2   Filter/557c7db83002419aa2634e0c0ca9f2e2/0  <none>      46 seconds ago      37 seconds ago      200 B
	
	# read the file
	$ pachctl get-file Sum fa59744b0e0448348159fef216f4eee9/0 apple
	133

.. note::

	If you're manually commiting new data, monitoring jobs to wait for them to finish, and then reading the latest commit in the output repo, you don't actually need flush-commit. But as soon as you have data streaming in or want to look up a result that corresponds to a specific input commit, ``flush-commit`` is your answer.

Check out the API docs for :doc:`../pachctl/pachctl_flush-commit` if you want a complete overview of the optional arguments.

