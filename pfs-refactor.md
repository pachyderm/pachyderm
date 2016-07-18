## Schema

### Notation

* `table` means a database table.
* `object` means a JSON object.
* `set<string>` can be simulated with map<string, bool> in JSON.
* `typeA | typeB` means either `typeA` or `typeB`.

### Schema

```
object Append {
  set<string> children;
  arr<string> blockrefs;
  bool delete;
}

object Clock {
  string branch;  // a branch ID
  int commit;
}

table Repo {
  string name;  // primary key
  Timestamp created;
}

table Branch {
  string ID; // repo + name
  string repo;
  string name; // a human readable name or a UUID

  index repoIndex = repo;
}

table Diff {
  string ID;  // commitID + path;
  string repo;
  string commit;
  string path;
  bool delete;  // If delete is true, the path has been deleted in this diff
  arr<Clock> clocks;  // the same as the commit's clocks
  arr<Append> appends;
  int size;

  index diffIndex = repo+delete+path+clocks;
}

table Commit {
  // A commit ID has the following format:
  //    repo/branch/integer
  // It's equivalent to the last component of clocks.
  // For instance, these are some valid commit IDs:
  //    data/foo/0
  //    data/foo/1
  //    data/bar/0
  string ID;  // primary key; 
  string repo;
  arr<Clock> clocks;
  Timestamp started;
  Timestamp finished;
  arr<string> provenance;  // commit IDs, topologically sorted

  index commitIndex = repo+clocks;
}
```

## Commits

### The Clock System

In Pachyderm, we use a [logical clock](https://en.wikipedia.org/wiki/Logical_clock) system to track the causal/parental relationships among commits.  The system is mostly equivalent to [vector clocks](https://en.wikipedia.org/wiki/Vector_clock), where each branch is the equivalent of a node in a distributed system, and each commit is the equivalent of an event.

Each commit carries a vector clock.  The rule for setting the vector clock is simple:

1. If the commit is the first commit on a new branch, we append a clock `(branch_name, 0)` to the vector.
2. Otherwise, the commit inherits its parent's vector clock, with the last component incremented by one.

For instance, this is a valid commit graph with three branches:

```
[(foo, 0)] -> [(foo, 1)] -> [(foo, 2)] -> [(foo, 2)]
           |
           -> [(foo, 0), (bar, 0)] -> [(foo, 0), (bar, 1)] -> [(foo, 0), (bar, 2)]
                                                           |
                                                           -> [(foo, 0), (bar, 1), (buzz, 0)]
```

Intuitively, commit A is commit B's ancestor if and only if:

1. Commit A's clock is a prefix of commit B's clock, or
2, Commit A's clock is the same as commit B's clock except that the last component is smaller.

Therefore, by creating a database index on the clocks of the commits, we can find all ancestors of a commit in one query. 

### Merging

Branches can be merged in two ways:

1. Replay: commits are copied onto the branch being merged to.
2. Squash: diffs are copied into a new commit on the branch being merged to.

## Code

### Notation

The code is all pseudocode.

* `Table.Operation` means a database operation on a table.
* `causedBy` is the vector clock comparison function.

### StartCommit(repoName, parentCommitID = null, branchName = null)

```
if parentCommit is null:
  if branchName is null:
    branchName = uuid()
  BranchTable.CreateIfNotExist(repoName+branchName, repoName, branch)
  commit = CommitTable.New()
  repo = RepoTable.Get(repoName)
  repo.BranchCounter += 1
  commit.Clocks += BranchClock(repo.BranchCounter, 0)
else:
  parentCommit = CommitTable.Get(parentCommitID)
  commit = parentCommit.Clone()
  if branch is null:
    commit.Clocks.lastElement.Clock += 1
  else:
    BranchTable.CreateIfNotExist(repoName+branchName, repoName, branch)
    repo = RepoTable.Get(repoName)
    repo.BranchCounter += 1
    commit.Clocks += BranchClock(repo.BranchCounter, 0)
return commit
```

### History(toCommit, fromCommit)

```
History(toCommit, fromCommit):
    clocksTo = CommitTable.Get(toCommit).Clocks
    clocksFrom = CommitTable.Get(fromCommit).Clocks
    return Between(clocksTo, clocksFrom)  // this can be done mathematically
```

### MergeCommits(repoName, fromCommits, parentCommitId = null, branch = null, squash = false)

### InspectFile(path, toCommit, fromCommit = null)

```
InspectFile(path, toCommit, fromCommit):
    clocksTo = CommitTable.Get(toCommit).Clocks
    clocksFrom = CommitTable.Get(fromCommit).Clocks
    // Find the most recent diff that removed the file
    newClocksFrom = diffIndex.Between(repo+true+path+clocksFrom, repo+true+path+clocksTo).filter(causedBy).first().Clocks
    return Between(clocksTo, newClocksFrom)  // start streaming blocks as soon as we get the first diff; can be done mathematically
```

### GetFile(commitID, path, fromCommitID)

same thing as InspectFile

### PutFile(commitID, path, reader)

```
blockrefs = read(reader)
diff = Diff.CreateIfNotExist(commitID, path)
diff.blockrefs.append(blockrefs)
```

### CreateJob

A job with a parallelism of N creates N branches based off the parent output commit.  Each shard operates on its own branch.  Once all shard finishes, we merge all branches into the original branch.

Note that since every shard operates on its own branch, there is no chance of blockrefs interleaving.  Therefore, we are able to do without "handles".

```
fromCommits = []
for shard in shards:
    commit = StartCommit(repo, parentOutputCommit)
    runJob(commit)
    fromCommits += commit
MergeCommits(repo, fromCommits, parentOutputCommit, squash = true)
```
