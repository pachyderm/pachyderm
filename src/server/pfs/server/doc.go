/*
Package server handles Pachyderm file system (PFS) requests.

PFS deals with repos, branches, commits, and provenance relationships between them.

# Provenance and Lineage

While our repos and commits are reasonably close to their git counterparts, branches in Pachyderm are more complicated.
Branches play an additional role of representing relationships between data in different repos.

This system is called provenance.

Provenance is a directional relationship (branch B is provenant on branch A if the contents of B depend on the contents of A).
Provenance information is stored in:
  - [pfs.BranchInfo.DirectProvenance], the branches which a branch directly depends upon.
    A pipeline's output branch has the spec branch as well as all input branches in direct provenance.
  - [pfs.BranchInfo.Provenance], the transitive closure of direct provenance.
  - [pfs.BranchInfo.Subvenance], the inverse of provenance (A is in B's provenance <=> B is in A's subvenance)

As a whole, provenance forms a directed, acyclic graph (DAG) - acyclic because data cannot depend on itself.

Provenance is what makes Pachyderm pipelines data-driven.
If branch B is provenant on branch A, we require that the head commit of B is provenant on the head commit of A.
New commits in A will cause new commits in B, which Pachyderm can use a signal to start new jobs to run pipeline code.

Commit provenance is similar to branch provenance, and is stored as a list of branches in [pfs.CommitInfo.DirectProvenance].
This list is locked in at creation time to match the commit's branch's provenance (all commits must be created on a branch).
We have an implicit rule that commit provenance only exists between commits with the same ID.
This simplifies accessing lineage information for commits, as a single ("global") ID refers to an entire chain of provenance-connected commits.

To support these relationships, we often create special commits which just associate existing commit data to a new ID.
These are alias commits, identified by [pfs.CommitInfo.Origin] having a kind of [pfs.OriginKind_ALIAS].
Alias commits cannot be modified directly, and instead inherit properties of their base commit (their closest non-alias ancestor).
Internally, when a commit is changed, we make sure to propagate the changes to any aliases, such as via driver.finishAliasDescendents.
Alias commits aren't generally important beyond their aliasing role, so they are often hidden by default from CLI commands.

# Commit States

A typical commit goes through three states, indicated by the timestamps of transitions between the states.

On creation, a commit is in the [pfs.CommitState_STARTED] state, during which its contents can be modified.
The timestamp [pfs.CommitInfo.Started] will be set for all commits.

After modifications are complete, the FinishCommit API endpoint moves a commit to [pfs.CommitState_FINISHING].
This sets the timestamp [pfs.CommitInfo.Finishing], and optionally [pfs.CommitInfo.Error].
Error indicates something is wrong with the commit (either it's file system is invalid, or it is the output of a failed job),
and in many places errored commits will be ignored entirely.
The naming can be confusing here, as to users the commit is "finished" after this point, and cannot be modified,
but internally there is more work to be done, and some operations will wait for a commit to be in the next state.

The PFS master collects "finishing" commits and finishes them asynchronously.
Already-errored commits will not be affected by this process other than setting [pfs.CommitInfo.Finished].
Typical commits will go through compaction, possibly consolidating the underlying file sets for more efficient access,
then a validation step which checks certain file system invariants which are necessary for proper access:
  - that no path is both a folder and a file
  - that no file is produced by more than one datum.

# Commit Contents (File Sets)

High level operations on commit contents interact with them like a normal file system.
The normal flow is to start a commit, add or delete files, then finish the commit.
Requests to modify the files at the head of a branch also automatically create commits with those changes.
Note that directory deletions can be expensive in Pachyderm,
as they require walking through the previous commit data to find files under the directory.

Internally, things are more complicated.
A commit's contents are represented by one or more file sets (see package [fileset]).
The mappings from commit to file sets are stored in two postgres tables:
  - pfs.commit_diffs
  - pfs.commit_totals

A [pfs.CommitState_FINISHED] commit has a single "total" file set and no diffs,
though it will usually be a [fileset.Composite] which contains others.
Otherwise, a commit will have an ordered list of one or more diff file sets, understood relative to a "base" ancestor commit.
The base commit is determined as part of driver.getFileSet as the most recent ancestor without an error.

The PFS driver modifies commit contents internally by adding or removing diff file sets.
The total file set is set once by compaction (see driver.finishRepoCommits), and never modified.
*/
package server
