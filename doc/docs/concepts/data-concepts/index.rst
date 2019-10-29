.. _data-concepts:

Versioned Data Concepts
=======================

Pachyderm data concepts describe version control primitives that
you interact with when you use Pachyderm.

These concepts are similar to the Git version control
system with a few notable exceptions. Because Pachyderm
deals not only with plain text but also with binary files and
large datasets, it does not manage the data in the same way as Git.
When you use Git, you store a copy of the repository on your
local machine. You work with that copy, apply your changes, and
then send the changes to the upstream master copy of the repository
where it gets merged.

The Pachyderm version control works slightly differently. In Pachyderm,
only a centralized repository exists, and you do not store any local copies
of that repository. Therefore, the merge, in the traditional Git meaning,
does not occur.

Instead, your data can be continuously updated in the master branch of
your repo, while you can experiment with specific data commits in a
separate branch or branches. Because of this behavior, you cannot
run into a merge conflict with Pachyderm.

The Pachyderm data versioning system has the following main concepts:

Repository
 A Pachyderm repository is the highest level data object. Typically,
 each dataset in Pachyderm is its own repository.

Commit
 A commit is an immutable snapshot of a repo at a particular point
 in time.

Branch
 A branch is an alias to a specific commit, or a pointer, that
 automatically moves as new data is submitted.

File
 Files and directories are actual data in your repository. Pachyderm
 supports any type, size, and number of files.

Provenance
 Provenance expresses the relationship between various
 commits, branches, and repositories. It helps you to track the origin
 of each commit.

Learn more about Pachyderm data concepts in the following sections:

.. toctree::
   :maxdepth: 1

   repo.md
   commit.md
   branch.md
   file.md
   provenance.md
   history.md
