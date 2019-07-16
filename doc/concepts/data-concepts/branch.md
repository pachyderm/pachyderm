# Branch

A Pachyderm branch is a pointer to a data snapshot. By default, Pachyderm
creates a `master` branch for each repository. Users can create additional
branches to experiment with the data and later merge these branches with
the master branch. Branching is a powerful feature that enables collaboration
between teams of data scientists. However, many users find it sufficient to
use the master branch for all their manipulations with data.

Each branch has a `HEAD` which references the latest commit in the
repository. By default, Pachyderm pipelines look at the `HEAD` of the branch
for changes and, if they detect new changes, launch. When you commit a new
change, the `HEAD` of the branch moves to the latest commit.

To view information about a branch, run the `pachctl list branch` command.

**Example:**

```bash
pachctl list branch images
BRANCH HEAD
master bb41c5fb83a14b69966a21c78a3c3b24
```
