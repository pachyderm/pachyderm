# Deleting Data in Pachyderm

Sometimes "bad" data gets committed to Pachyderm and you need a way to delete
it. There are a couple of ways to address this, which depend on
what exactly was "bad" about the data you committed and what's happened in the
system since you committed the "bad" data.

- [Deleting the HEAD of a branch](#deleting-the-head-of-a-branch) - You should
follow this guide if you've just made a commit to a branch with some corrupt, incorrect,
or otherwise bad changes to your data.
- [Deleting non-HEAD commits](#deleting-non-head-commits) - You should follow
this guide if you've committed data to the branch after committing the data that
needs to be deleted.
- [Deleting sensitive data](#deleting-sensitive-data) - You should follow these
steps when you have committed sensitive data that you need to completely
purge from Pachyderm, such that no trace remains.

## Deleting The HEAD of a Branch

The simplest case is when you've just made a commit to a branch with some
incorrect, corrupt, or otherwise bad data. In this scenario, the HEAD of your branch
(i.e., the latest commit) is bad. Users who read from it are likely to be misled, and/or
pipeline subscribed to it are likely to fail or produce bad downstream output.

To fix this you should use `delete-commit` as follows:

```sh
$ pachctl delete-commit <repo> <branch-or-commit-id>
```

When you delete the bad commit, several things will happen (all atomically):

- The commit metadata will be deleted.
- Any branch that the commit was the HEAD of will have its HEAD set to the
  commit's parent. If the commit's parent is `nil`, the branch's HEAD will be set
  to `nil`.
- If the commit has children (commits which it is the parent of), those
  children's parent will be set to the deleted commit's parent. Again, if the
  deleted commit's parent is `nil` then the children commit's parent will be
  set to `nil`.
- Any jobs which were created due to this commit will be deleted (running jobs
  get killed). This includes jobs which don't directly take the commit as
  input, but are farther downstream in your DAG.
- Output commits from deleted jobs will also be deleted, and all the above
  effects will apply to those commits as well.

## Deleting Non-HEAD Commits

Recovering from commits of bad data is a little more complicated if you've
committed more data to the branch after the bad data was added. You can
still delete the commit as in the previous section, however, unless the subsequent
commits overwrote or deleted the bad data, it will still be present in the
children commits. *Deleting a commit does not modify its children.*

In git terms, `delete-commit` is equivalent to squashing a commit out of existence.
It's not equivalent to reverting a commit. The reason for this behavior is that the
semantics of revert can get ambiguous when the files being reverted have been
otherwise modified. Git's revert can leave you with a merge conflict to solve,
and merge conflicts don't make sense with Pachyderm due to the shared nature of
the system and the size of the data being stored.

In these scenario, you can also delete the children commits, however those commits
may also have good data that you don't want to delete. If so, you should:

1. Start a new commit on the branch with `pachctl start-commit`.
2. Delete all bad files from the newly opened commit with `pachctl delete-file`.
3. Finish the commit with `pachctl finish-commit`.
4. Delete the initial bad commits and all children up to the newly finished
   commit.

Depending on how you're using Pachyderm, the final step may be optional. After
you finish the "fixed" commit, the HEADs of all your branches will converge to
correct results as downstream jobs finish. However, deleting those commits
allow you to clean up your commit history and makes sure that no one will ever
access errant data when reading non-HEAD version of the data.

## Deleting Sensitive Data

If the data you committed is bad because it's sensitive and you want to make
sure that nobody ever accesses it, you should complete an extra step in addition to those
above.

Pachyderm stores data in a content addressed way and when you delete
a file or a commit, Pachyderm only deletes references to the underlying data, it
doesn't delete the actual data until it performs garbage collection. To truly
purge the data you must delete all references to it using the methods described
above, and then you must run a garbage collect with `pachctl garbage-collect`.
