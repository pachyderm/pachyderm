## Deferred Processing of Data

While they're running, Pachyderm Pipelines will process any new data you
commit to their input branches. This can be annoying in cases where you
want to commit data more frequently than you want to process.

This is generally not an issue because Pachyderm pipelines are smart about
not reprocessing things they've already processed, but some pipelines need
to process everything from scratch. For example, you may want to commit
data every hour, but only want to retrain a machine learning model on that
data daily since it needs to train on all the data from scratch.

In these cases there's a massive performance benefit to deferred
processing. This document covers how to achieve that and control exactly
what gets processed when using the Pachyderm system.

The key thing to understand about controlling when data is processed in
Pachyderm is that you control this using the _filesystem_, rather than at
the pipeline level. Pipelines are inflexible but simple, they always try
to process the data at the heads of their input branches. The filesystem,
on the other hand, is much more flexible and gives you the ability to
commit data in different places and then efficiently move and rename the
data so that it gets processed when you want. The examples below describe
how specifically this should work for common cases.

### Using a staging branch

The simplest and most common pattern for deferred processing is using
a `staging` branch in addition to the usual `master` branch that the
pipeline takes as input. To begin, create your input repo and your
pipeline (which by default will read from the master branch). This will
automatically create a branch on your input repo called `master`. You can
check this with `list-branch`:

```sh
$ pachctl list-branch data
BRANCH HEAD
master -
```

Notice that the head commit is empty. This is why the pipeline has no jobs
as pipelines process the HEAD commit of their input branches. No HEAD
commit means no processing. If you were to commit data to the `master`
branch, the pipeline would immediately kick off a job to process what you
committed. However, if you want to commit something without immediately
processing it you need to commit it to a different branch. That's where
a `staging` branch comes in -- you're essentially adding your data into
a staging area to then process later.

Commit a file to the staging branch:

```sh
$ pachctl put-file data staging -f <file>
```

Your repo now has 2 branches, `staging` and `master` (`put-file`
automatically creates branches if they don't exist). If you do
`list-branch` again you should see:

```sh
$ pachctl list-branch data
BRANCH  HEAD
staging f3506f0fab6e483e8338754081109e69
master  -
```

Notice that `master` still doesn't have a head commit, but the new branch,
`staging`, does. There still have been no jobs, because there are no
pipelines taking `staging` as inputs. You can continue to commit to
`staging` to add new data to the branch and it still won't process
anything. True to its name, it's acting as a staging ground for data.

When you're ready to actually process the data all you need to do is
update the master branch to point to the head of the staging branch:

```sh
$ pachctl create-branch data master --head staging
$ pachctl list-branch
staging f3506f0fab6e483e8338754081109e69
master  f3506f0fab6e483e8338754081109e69
```

Notice that `master` and `staging` now have the same head commit. This
means that your pipeline finally has something to process. If you do
`list-job` you should see a new job. Notice that even if you created
multiple commits on `staging` before updating `master` you still only get
1 job. Despite the fact that those other commits are ancestors of the
current HEAD of master, they were never the actual HEAD of `master`
themselves, so they don't get processed. This is often fine because
commits in Pachyderm are generally additive, so processing the HEAD commit
also processes data from previous commits.

![deffered processing](deferred_processing.gif)

However, sometimes you want to
process specific intermediary commits. To do this, all you need to do is
set `master` to have them as HEAD. For example if you had 10 commits on
`staging` and you wanted to process the 7th, 3rd, and most recent
commits, you would do:

```sh
$ pachctl create-branch data master --head staging^7
$ pachctl create-branch data master --head staging^3
$ pachctl create-branch data master --head staging
```

If you do `list-job` while running the above commands, you will see
between 1 and 3 new jobs. Eventually there will be a job for each of the
HEAD commits, however Pachyderm won't create a new job until the previous
job has completed.

#### What to do if you accidentally process something you didn't want to

In all of the examples above we've been *advancing* the `master` branch to
later commits. However, this isn't a requirement of the system, you can
move backward to previous commits just as easily. For example, if after
the above commands you realized that actually want your final output to be
the result of processing `staging^1`, you can "roll back" your HEAD commit
the same way we did before.

```sh
$ pachctl create-branch data master --head staging^1
```

This will kick off a new job to process `staging^1`. The HEAD commit on
your output repo will be the result of processing `staging^1` instead of
`staging`.

### More complicated staging patterns

Using a `staging` branch allows you to defer processing, but it's
inflexible, in that you need to know ahead of time what you want your
input commits to look like. Sometimes you want to be able to commit data
in an ad-hoc, disorganized way and then organize it later. For this,
instead of updating your `master` branch to point at commits from
`staging`, you can copy files directly from `staging` to `master`. With
`copy-file`, this only copies references, it doesn't move the actual data
for the files around.

This would look like:

```sh
$ pachctl start-commit data master
$ pachctl copy-file data staging file1 data master
$ pachctl copy-file data staging file2 data master
...
$ pachctl finish-commit data master
```

You can also, of course, issue `delete-file` and `put-file` while the commit is
open if you want to remove something from the parent commit or add something
that isn't stored anywhere else.
