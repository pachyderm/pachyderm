# Repository

A Pachyderm repository is a location where you store your data inside
Pachyderm. A Pachyderm repository is a top-level data object that contains
files and folders. Similarly to Git, a Pachyderm repository tracks all
changes to the data and creates a history of data modifications that you
can access and review. You can store any type of file in a Pachyderm repo,
including binary and plain text files.

Unlike a Git repository that stores history in a `.git` file in your copy
of a Git repo, Pachyderm stores the history of your commits in a centralized
location in the etcd key-value store. Because of that, you do not run into
merge conflicts as you often do with Git commits when you try to merge
your `.git` history with the master copy of the repo. With large datatsets
resolving a merge conflict might not be possible.

A Pachyderm repository is the first entity that you configure to create
a standard pipeline. You can create a repository by running the
`pachctl create repo <name>` command or by using the Pachyderm UI. After
creating the repository, you can add your data by using the
`pachctl put file <repo@branch:/path><path-to-filesystem>` command. The
path to the filesystem can be a local directory or file or a URL.
This repository is called the *input repo* and is stored under
`pfs/<repo-name>`. All the input data for your pipeline is stored in
this repository.

After you create a repository, you can specify that repository as a
PFS input in your pipeline. When the pipeline detects
new unprocessed data in the repository, it runs your code against it.

For each pipeline, Pachyderm automatically creates a `pfs/out` directory
in the `pachd` container. This directory is called the *output repository*.
All your output data is stored in this directory.
