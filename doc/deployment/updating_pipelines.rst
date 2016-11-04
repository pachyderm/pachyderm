Updating Pipelines
==================



During development, it's very common to update pipelines, whether it's changing your code or just cranking up the parallelism. 

In the former case, changing your pipeline code completely invalidates previous results in the output repo and breaks parantage of output commits for this pipeline and all downstream pipelines. By default, Pachyderm won't delete the data from those downstream repos, instead we ``archive`` it so you can still access the data and we then reprocess the input data again with your new code. 

.. note::

	The update-pipeline functionality is somewhat in beta and the API will likely change a bunch in the next release. We'd love your feedback an how you're using it and what we can improve. info@pachyderm.io. 

.. code-block:: shell
	
	$ pachctl update-pipeline -f pipeline.json 

In some cases, such as changing parallelism, where you don't want to archive previous data and recompute results, you can pass the ``-no-archive`` flag.

.. warning::

	To emphasize again, updating a pipeline without the -no-archive flag will archive ALL downstream repos and start reprocessing ALL the input commits again. 0


Archived commits
----------------
Archived commits are meant for old data that isn't relevant to the current set of pipelines but you might want to reference later. 

Archived commits are nearly identical to regular commits except they don't show up on ``list-commit`` or ``flush-commit``. You can use ``list-commit -a`` to see both archived and canceled commits and can still read from archived commits by referencing them by ``commitID``.

You can also manually archive data:

.. code-block:: shell

	# Archive all commits in all repos. This is ideal if you have a bunch of garbage data, but want
	# to keep your pipelines and repos intact. Your old data is still available using list-commit -a.
	$ pachctl archive-all

	# Archive a specific commit. 
	$ archive-commit <commitID>

.. note::
	
	Only use ``archive-commit`` if you know what you're doing, or archiving a whole branch, because it can get you into some weird situations. 









