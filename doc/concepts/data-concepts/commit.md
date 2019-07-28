# Commit

A commit is a snapshot that preserves the state of your data at a point in time.
It represents a single change to a file or files, directory or directories
in your Pachyderm repository. Every time you modify or add new data to a
Pachyderm input repository, Pachyderm detects and records it as a new
commit or revision.

Each commit has a unique identifier (ID) that you can reference in
the `<repo-name@commit-number>` format. When you create a new
commit, the previous commit on which the new commit is based becomes
the parent of the new commit.
The Directed acyclic graph (DAG) that represents your pipeline history
consists of those parent-child relationships between your data commits.

You can obtain information about commits in a repository by running `list commit <repo>` or `inspect commit <commitID>`
the `pachctl list commit repo@branch` command. This command returns a
timestamp, size, parent, and other information about the commit.
The initial commit has `<none>` as a parent.

**Example:**

```bash
pachctl list commit images@master
REPO   BRANCH COMMIT                           PARENT                           STARTED        DURATION           SIZE
images master 8e8eb2bf46d449f18117bd5e24479e43 b7cbe0445b2342e98b8ffcde87029d1a 11 seconds ago Less than a second 255.9KiB
images master b7cbe0445b2342e98b8ffcde87029d1a <none>                           23 hours ago   Less than a second 238.3KiB
```

You can subscribe your branches to specific commits and run your code
against the data in these specific commits. Smaller changes can be combined
into a single transaction by using the `start commit` and `finish commit`
commands.

The `flush commit` command that enables you to
see in which outputs the commit was used. Therefore, you can track
commit provenance downstream, from the commit origin to the final result.
