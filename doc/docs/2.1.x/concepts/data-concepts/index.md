# Pachyderm Versioned Data Concepts

Pachyderm data concepts describe **version-control primitives** that
you interact with when you use Pachyderm.

These ideas are conceptually similar to the Git version-control
system with a few notable exceptions. Because Pachyderm
deals not only with plain text but also with binary files and
large datasets, it does not process the data in the same way as Git.
When you use Git, you store a copy of the repository on your
local machine. You work with that copy, apply your changes, and
then send the changes to the upstream master copy of the repository
where it gets merged.

Pachyderm version control works slightly differently. In Pachyderm,
only a centralized repository exists and you do not store any local copies
of that repository. Therefore, the merge, in the traditional Git meaning,
does not occur.

Instead, your data can be continuously updated in the master branch of
your repo, while you can experiment with specific data commits in a
separate branch or branches. Because of this behavior, you cannot
run into a merge conflict with Pachyderm.

The Pachyderm data versioning system has the following main concepts:

## **Repository**
A [Pachyderm repository](./repo/) is the highest level data object,
an independent file system. Typically, each dataset in
Pachyderm is its own repository. 

## **Commit**
A [commit](./commit/) is an immutable snapshot of a data in Pachyderm corresponding to a change in source data or transformations. Unlike in git, a commit records the state of a branche in a repository.

## **Branch**
A [branch](./branch/) is the basic unit of provenance. At any given time, it points
to the state of its repo at a particular commit, updating as new data
is added. Crucially, a branch also tells Pachyderm what input it depends on.

## **File**
[Files](./file/) are actual data in your repository. Pachyderm
supports any type, size, and number of files.

## **Provenance**
[Provenance](./provenance/) expresses the relationship between branches in different
repositories. It helps you understand the origin of all data in Pachyderm,
ensuring each commit contains enough information to reconstruct the past.

