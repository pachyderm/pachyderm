Analyze Your Data
=================

This section of documentation covers everything you'll need to know to deploy a working Pachyderm cluster and build your own analysis to process whatever data you want. 

If you're brand new to Pachyderm, you should check out our :doc:`../getting_started/getting_started` documentation to install Pachyderm locally and learn the basic concepts. 

:doc:`deploying_on_the_cloud`: Get Pachyderm deployed on AWS, GCE, or OpenShift.

:doc:`inputing_your_data`: Get your own data into Pachyderm.

:doc:`custom_pipelines`: Get your code running in Pachyderm and processing data.

:doc:`pipeline_spec`: A complete reference on the advanced features of Pachyderm pipelines.

:doc:`getting_your_results`: Read out results from specific input commits.

:doc:`updating_pipelines`: Interate on pipelines as you learn from your data.


Usage Metrics
-------------

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.



