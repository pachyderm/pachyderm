.. Pachyderm documentation master file, created by
   sphinx-quickstart on Thu Jul  7 10:45:21 2016.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

.. _Go Client: https://godoc.org/github.com/pachyderm/pachyderm/src/client

Pachyderm Developer Documentation
=================================

User Documentation
------------------

* :ref:`first_time_section`
* :ref:`development_section`
* :ref:`workflows_section`
* :ref:`production_section`
* :ref:`reference-section`

.. _first_time_section:

.. toctree::
    :maxdepth: 2
    :caption: First Time Users

    getting_started

.. _development_section:

.. toctree::
    :maxdepth: 2
    :caption: Analyze Your Data

    development/deploying_on_the_cloud
    development/inputing_your_data
    development/customizing_your_images
    development/updating_pipelines
    development/scaling
    development/cluster_hygiene
    development/python_logs_example

.. _workflows_section:

.. toctree::
    :maxdepth: 2
    :caption: Advanced Workflows

    workflows

.. _production_section:

.. toctree::
    :maxdepth: 2
    :caption: Going To Production

    production


.. _reference-section:

.. toctree::
    :maxdepth: 2
    :caption: Reference By Feature

    pachyderm_file_system
    pachyderm_pipeline_system
    pachctl
    `Go Client`_ #TODO - this link is missing from the list
    examples
    CHANGELOG


