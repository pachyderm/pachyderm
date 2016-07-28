.. Pachyderm documentation master file, created by
   sphinx-quickstart on Thu Jul  7 10:45:21 2016.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

.. _Go Client: https://godoc.org/github.com/pachyderm/pachyderm/src/client

Pachyderm Developer Documentation
=================================

* :ref:`user_docs`
* :ref:`feature_docs`

.. _user_docs:

.. toctree::
    :maxdepth: 2
    :caption: First Time Users

    getting_started

.. toctree::
    :maxdepth: 2
    :caption: Analyze Your Data

    development/deploying_on_the_cloud
    development/inputing_your_data
    development/customizing_your_images
    development/tracking_your_data
    development/updating_pipelines
    development/scaling
    development/cluster_hygiene
    development/python_logs_example

.. toctree::
    :maxdepth: 2
    :caption: Advanced Workflows

    advanced/provenance
    advanced/incrementality
    advanced/complex_dags

.. toctree::
    :maxdepth: 2
    :caption: Going To Production

    production


.. _feature_docs:

.. toctree::
    :maxdepth: 2
    :caption: Reference By Feature

    pachyderm_file_system
    pachyderm_pipeline_system
    pachctl
    `Go Client`_ #TODO - this link is missing from the list
    examples
    CHANGELOG


