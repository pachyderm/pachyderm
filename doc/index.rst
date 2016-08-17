.. Pachyderm documentation master file, created by
   sphinx-quickstart on Thu Jul  7 10:45:21 2016.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

.. _Go Client: https://godoc.org/github.com/pachyderm/pachyderm/src/client

Pachyderm Developer Documentation
=================================

:ref: `user_docs`
:ref: `feature_docs`
:ref: `development/development`
:ref: `development`



.. _user_docs:

.. toctree::
    :maxdepth: 2
    :caption: Getting Started

    getting_started/getting_started
    getting_started/local_installation
    getting_started/beginner_tutorial
    getting_started/troubleshooting

.. toctree::
    :maxdepth: 2
    :caption: Analyze Your Data
    
    development/analyze_your_data
    development/deploying_on_the_cloud
    development/inputing_your_data
    development/customizing_your_images
    development/tracking_your_data
    development/updating_pipelines
    development/scaling
    development/cluster_hygiene
    development/python_example

.. toctree::
    :maxdepth: 2
    :caption: Advanced Workflows

    advanced/advanced
    advanced/provenance
    advanced/incrementality
    advanced/composing_pipelines

.. toctree::
    :maxdepth: 2
    :caption: Going To Production

    production/production
    production/collaboration
    production/connecting_to_other_data_sources
    production/outputting_results
    production/best_practices
    production/performance_benchmarks
    production/maintenance


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


