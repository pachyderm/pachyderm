Analyze Your Data
=================

This section of documentation covers everything you'll need to know to deploy a working Pachyderm cluster and build your own analysis to process whatever data you want. 

If you're brand new to Pachyderm, you should check out our :doc:`getting_started` documentation to install Pachyderm locally and learn the basic concepts. 

.. toctree::
    :maxdepth: 1
    
    deploying_on_the_cloud
    inputing_your_data
    custom_pipelines
    tracking_your_data
    updating_pipelines
    scaling
    cluster_hygiene
    python_example


Usage Metrics
-------------

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.
