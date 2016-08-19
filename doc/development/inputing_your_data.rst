Getting Your Data into Pachyderm
================================

If you're running Pachyderm in the cloud, data in Pachyderm is backed an object store such as S3 or GCS. Files in Pachyderm are content-addressed as part of how we buid our version control semantics and are therefore not "human-readable." We recommend you give Pachyderm its own bucket.

There a bunch of different ways to get your data into Pachyderm.

`PFS Mount`_: This is the easiest method if you just have some local files (or dummy files) and you just want to test things out in Pachyderm.

`Pachtl CLI`_: This is the best option for real use cases and scripting the input process.

`Golang Client`_: Ideal for Golang users who want to script the file input process.

`Other Language Clients`_: Pachyderm uses a protocol buffer API which supports many other languages, we just haven't built full clients yet. 




PFS  Mount
----------

To mount pfs locally use ``pachctl mount``. You can follow the first section of the :doc:`beginner_tutorial` or use the :doc:`pachctl` documentation.

Once you have pfs mounted, you can add files to Pachyderm via whatever method you prefer to manipulate a local file system:  ``mv``, ``cp``, ``>``, ``|``, etc.

Don't forget, you'll need create a repo in Pachyderm first with ``pachctl create-repo <repo_name>`` and add the files to ``~/pfs/<repo_name>``.


Pachctl CLI
-----------

The pachctl CLI is the primary method of interaction with Pachyderm. To get data into Pachyderm, you should use the ``put-file`` command. Below are a examples uses of ``put-file``. Go to :doc:`pachctl/pachctl_put-file`` for complete documentation. 

.. note::

  Commits in Pachyderm must be explicitly started and finished so put-file can only be called on an open (started, but not finished) commit. The ``-c`` option allows you to start and finish the commit in addition to putting data as a one-line command. 


-c to start and finish put-file as a oneliner. 

Adding a single file:

Adding multiple files at once:

Scraping a URL:




2) You can do it via the [pachctl](./pachctl.html) CLI

This is probably the most reasonable if you're scripting the input process.

3) You can use the golang client

You can find the docs for the client [here](https://godoc.org/github.com/pachyderm/pachyderm/src/client)

Again, helpful if you're scripting something, and helpful if you're already using go.




Golang Client
-------------

Other Language Clients
----------------------


Data Sources
------------


From Local Files
^^^^^^^^^^^^^^^^

From your Object Storage
^^^^^^^^^^^^^^^^^^^^^^^^

From a Database
^^^^^^^^^^^^^^^

One-time dump

Ongoing sync with a DB. 