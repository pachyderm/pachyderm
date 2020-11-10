# Repository

A Pachyderm repository is a location where you store your data inside
Pachyderm. A Pachyderm repository is a top-level data object that contains
files and folders. Similar to Git, a Pachyderm repository tracks all
changes to the data and creates a history of data modifications that you
can access and review. You can store any type of file in a Pachyderm repo,
including binary and plain text files.

Unlike a Git repository that stores history in a `.git` file in your copy
of a Git repo, Pachyderm stores the history of your commits in a centralized
location. Because of that, you do not run into
merge conflicts as you often do with Git commits when you try to merge
your `.git` history with the master copy of the repo. With large datatsets
resolving a merge conflict might not be possible.

A Pachyderm repository is the first entity that you configure when you want
to add data to Pachyderm. You can create a repository with the `pachctl create repo`
command, or by using the Pachyderm UI. After creating the repository, you can
add your data by using the `pachctl put file` command.

A Pachyderm repo name can include alphanumeric characters, dashes, and underscores,
and should be no more than 63 characters long.

The following types of repositories exist in Pachyderm:

Input repositories
:   Users or external applications outside of Pachyderm can add data to
    the input repositories for further processing.

Output repositories
:   Pachyderm automatically creates output repositories
    pipelines write results of computations into these repositories.

You can view the list of repositories in your Pachyderm cluster
by running the `pachctl list repo` command.

!!! example
    ```shell
    pachctl list repo
    ```

    **System Response:**

    ```shell
    NAME     CREATED     SIZE (MASTER)
    raw_data 6 hours ago 0B
    ```

The `pachctl inspect repo` command provides a more detailed overview
of a specified repository.

!!! example
    ```shell
    pachctl inspect repo raw_data
    ```

    **System Response:**

    ```shell
    Name: raw_data
    Description: A raw data repository
    Created: 6 hours ago
    Size of HEAD on master: 5.121MiB
    ```

If you need to delete a repository, you can run the
`pachctl delete repo` command. This command deletes all
data and the information about the specified
repository, such as commit history. The delete
operation is irreversible and results in a
complete cleanup of your Pachyderm cluster.
If you run the delete command with the `--all` flag, all
repositories will be deleted.

!!! note "See Also:"
    [Pipeline](../pipeline-concepts/pipeline/index.md)
