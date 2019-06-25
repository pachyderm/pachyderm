.. _concepts:

Concepts
========

Pachyderm enables you to create data pipelines that enable you
continuously track changes that data scientists, developers, or other
members of your team make to datasets. In other words, Pachyderm
is a version-control platform for your data similar to what Git is
for source code.

Data scientints and researches often make changes to
data in batches and phases. Often, they need to go back in time to
revisit previous results or combine these data results with other
data in different variants. Pachyderm helps solving the issues
with reproducibility and data provenance by creating data pipelines
which store the history of your changes from the begining of times.

The two main Pachyderm concepts that
encapsulate building blocks of the Pachyderm solution are
Pachyderm File System (PFS) and Pachyderm Pipeline System (PPS).
You need to understand the basics of these concepts to start working
with Pachyderm and later might need to dive even deeper to the
more advanced technical topics to run more sophisticated Pachyderm
scenarios.

This section describes main Pachyderm concepts including the following
topics:

.. toctree::
   :maxdepth: 2

   pfs/index.rst
   pps/index.rst

