Troubleshooting
===============

Below a list of common errors we run accross with users trying to run Pachyderm locally and following the :doc:`beginner_tutorial`. 

One of the first things you should do is check the logs:

.. code-block:: shell

 $ pachctl list-job
 $ get-logs <jobID>

If the problem is with deployment of Pachyderm, Kubernetes has their own logs too:

.. code-block:: shell

	$ kubectl get all
	$ kubectl logs pod_name 


**Error**: error messages with "cannot unmarshal"
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

**Solution**: This is usually due to a version mismatch. Start with ``pachctl version`` and make sure your client and server version are matching. 

If you got the error when running a pipeline, it's likely you're using the wrong version of the pipeline spec. For example, ``json: cannot unmarshal bool into Go value of type pps.Incremental`` is because the pipeline spec between v1.1 and v1.2 changed the type for incrememental from ``bool`` to ``string``. Refer to :doc:`../development/pipeline_spec` to check that yours is correct.


**Error**: Job status is stuck in "pulling" state
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

**Solution**: This means that Pachyderm can't find the image specified in your pipeline. It's possible you have a typo in your pipeline spec. More likely, the image isn't available on DockerHub, your DOcker registry, or locally to the Pachyderm pods. If you're running Pachyderm in Minikube, you need ot make sure any images you've built locally are accessible to Pachyderm within the VM. 


**Error**: ``pipeline <NAME> already exists``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
or ``repo <NAME> already exists``

**Solution**: Pipelines and repos need to have unique names. The errors happens most commonly when your first attempt at creating or running a pipeline failed and you try to do ``create-pipeline`` again. There are a bunch of solutions. 

	1. Delete the previous pipelines and associated repos. ``pachctl delete-repo <name>`` and  ``pachctl delete-pipeline <name>``.
	2. Change the "name" field in your pipeline spec JSON file. 
	3. Use the ``update-pipeline`` command. ``update-pipeline`` is a bit more complicated to use so make sure to read the :doc:`../development/updating_pipelines` docs.


**Error**: A pipeline using ``grep`` fails with a [1] error code.
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

**Solution**: Little known fact, ``grep`` errors with [1] if it does not find any results. For our fruit stand example, if we did grepped for grapes, the pipeline would end up failing. It's really easy to solve this by adding "acceptReturnCode": [1] in stdin of your pipeline spec. 

Technically, this [1] error could be coming from some other part of your code, but the most common culprit we've seen if you're running simple examples is grep.


**Error**: SOMETHING IS F***ED AND I WANT TO START OVER!
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

**Solution**: We got your back!

.. code-block:: shell

	# Delete all pipelines and repos in Pachyderm
	$ pachctl delete-all

	# Archive all commits in all repos. This is ideal if you have a bunch of garbage data, but want
	# to keep your pipelines and repos intact. Your old data is still available using list-commit -a.
	$ pachctl archive-all

	# Remove Pachyderm from your kubernetes cluster
	$ pachctl deploy --dry-run | kubectl delete -f -

	# Kill the entire minikube VM and restart. Don't skip the minikube delete step
	# because it keeps around some weird intermediate state.
	$ minikube stop
	$ minikube delete
	$ minikube start
