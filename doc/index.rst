.. Pachyderm documentation master file, created by
   sphinx-quickstart on Thu Jul  7 10:45:21 2016.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Pachyderm Developer Documentation
=================================

.. _Go Client: https://godoc.org/github.com/pachyderm/pachyderm/src/client
.. Example Links:
.. _Fruit Stand: https://github.com/pachyderm/pachyderm/tree/master/examples/fruit_stand
.. _Web Scraper: https://github.com/pachyderm/pachyderm/tree/master/examples/scraper
.. _Tensor Flow: https://github.com/pachyderm/pachyderm/tree/master/examples/tensor_flow
.. _Word Count: https://github.com/pachyderm/pachyderm/tree/master/examples/word_count

Getting Started
---------------

Already got a kubernetes cluster? Use it to create a pachyderm cluster:

```sh
$ kubectl create -f https://pachyderm.io/manifest.json
```

(You'll still need to install :doc:`pachctl`, our CLI)

If you don't have a kubernetes cluster setup, check out our :doc:`installation` instructions.

If you've never used Pachyderm before you should look at the `Fruit Stand`_ example. 

__


User Documentation
******************

* :doc:`about`
* :doc:`installation`
* :ref:`deploying_section`
* :ref:`examples_section`
* :doc:`troubleshooting`
* :doc:`FAQ`

Feature Documentation
*********************

* :doc:`pachyderm_file_system`
* :doc:`pachyderm_pipeline_system`
* :doc:`pachctl`
* `Go Client`_
* :doc:`CHANGELOG`

.. _deploying_section:

.. toctree::
    :maxdepth: 3
    :caption: Deploying

    deploying

.. _examples_section:

.. toctree::
    :maxdepth: 1
    :caption: Examples

- `Fruit Stand`_
- `Web Scraper`_
- `Word Count`_
- `Tensor Flow`_


